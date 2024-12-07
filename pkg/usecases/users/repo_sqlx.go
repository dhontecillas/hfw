package users

import (
	"crypto/sha256"
	"database/sql"
	"encoding/hex"
	"fmt"
	"math/rand"
	"time"

	"github.com/jmoiron/sqlx"
	"golang.org/x/crypto/bcrypt"

	"github.com/dhontecillas/hfw/pkg/db"
	"github.com/dhontecillas/hfw/pkg/ids"
	"github.com/dhontecillas/hfw/pkg/obs"
)

var _ RegistrationRepo = (*RepoSQLX)(nil)

// RepoSQLX implemnte the RegistrationRepo interface
// with a SQL db.
type RepoSQLX struct {
	sqlDB     db.SQLDB
	ins       *obs.Insighter
	tokenSalt string
}

// NewRepoSQLX creates a new RepoSQLX
func NewRepoSQLX(ins *obs.Insighter, sqlDB db.SQLDB,
	tokenSalt string) *RepoSQLX {
	return &RepoSQLX{
		sqlDB:     sqlDB,
		ins:       ins,
		tokenSalt: tokenSalt,
	}
}

func (r *RepoSQLX) scanUser(row *sqlx.Row, u *User) error {
	var idStr string
	var email string
	var created time.Time
	if err := row.Scan(&idStr, &email, &created); err != nil {
		return err
	}
	if err := u.ID.FromUUID(idStr); err != nil {
		return err
	}
	u.Email = email
	u.Created = created
	return nil
}

func (r *RepoSQLX) getUserByID(tx *sqlx.Tx, id ids.ID) *User {
	userQ := `
SELECT
	id
	,email
	,created
FROM users
WHERE
	id = $1
`
	strUID := id.ToUUID()
	row := tx.QueryRowx(userQ, id.ToUUID())
	var u User
	err := r.scanUser(row, &u)
	if err != nil {
		if err != sql.ErrNoRows {
			r.ins.L.Err(err, "cannot find user id", map[string]interface{}{
				"query": userQ,
				"id":    strUID,
			})
		}
		return nil
	}
	return &u
}

func (r *RepoSQLX) getUserByEmail(tx *sqlx.Tx, email string) *User {
	userQ := `
SELECT
	id
	,email
	,created
FROM users
WHERE
	email = $1
`
	row := tx.QueryRowx(userQ, email)
	var u User
	err := r.scanUser(row, &u)
	if err != nil {
		if err != sql.ErrNoRows {
			r.ins.L.Err(err, "cannot find user from email", map[string]interface{}{
				"query": userQ,
				"email": email,
			})
		}
		return nil
	}
	return &u
}

// GetUserByEmail returns the User for a given email,
// or nil if the email is not found in the data repo.
func (r *RepoSQLX) GetUserByEmail(email string) *User {
	// we do not need a transaction here, but since other getUserByID
	// works with a slqx.Tx param, we reuse it here
	master := r.sqlDB.Master()
	tx, _ := master.Beginx()
	u := r.getUserByEmail(tx, email)
	_ = tx.Commit()
	return u
}

// GetUserByID returns a User for a given ID, or nil
// if there is no user for that ID.
func (r *RepoSQLX) GetUserByID(userID ids.ID) *User {
	// we do not need a transaction here, but since other getUserByID
	// works with a slqx.Tx param, we reuse it here
	master := r.sqlDB.Master()
	tx, _ := master.Beginx()
	u := r.getUserByID(tx, userID)
	_ = tx.Commit()
	return u
}

// createToken creates a
func (r *RepoSQLX) createToken(email string) string {
	unhashedToken := fmt.Sprintf("%s%d%s%d", r.tokenSalt,
		rand.Uint64(), email, time.Now().UnixNano())
	token := sha256.Sum256([]byte(unhashedToken))
	return hex.EncodeToString(token[:])
}

func (r *RepoSQLX) passwordHash(password string) string {
	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		panic("cannot hash password")
	}
	return string(hash)
}

func (r *RepoSQLX) checkPassword(password string, hashedPassword string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(hashedPassword), []byte(password))
	return err == nil
}

type registrationRequest struct {
	token     string
	email     string
	requested time.Time
	expires   time.Time
	password  string
	consumed  time.Time
}

// CreateInactiveUser return a token for the user to be used
// to confirm the account.
func (r *RepoSQLX) CreateInactiveUser(
	email string, password string) (string, error) {

	now := time.Now()
	master := r.sqlDB.Master()
	tx, err := master.Beginx()
	if err != nil {
		return "", err
	}
	defer func() { _ = tx.Commit() }()

	u := r.getUserByEmail(tx, email)
	if u != nil {
		if rbErr := tx.Rollback(); rbErr != nil {
			r.ins.L.Err(rbErr, "rollback failed", nil)
		}
		r.ins.L.Err(ErrUserExists, "email already exists", map[string]interface{}{
			"email": email,
		})
		return "", ErrUserExists
	}

	// to discard existing activations, we set the expiration date
	// to the same requested date
	discardExistingRequestQ := `
UPDATE user_registration_requests
SET
	expires = requested
WHERE
	email = $1
	AND consumed IS NULL
`
	if _, err := tx.Exec(discardExistingRequestQ, email); err != nil {
		// just log the error
		r.ins.L.Err(err, " cannot update registration requests", nil)
	}

	hashedPass := r.passwordHash(password)
	token := r.createToken(email)
	expirationHours := 24
	expires := now.Add(time.Duration(expirationHours) * time.Hour)
	sqlQ := `
INSERT INTO user_registration_requests(
	email
	,token
	,requested
	,expires
	,password
)
VALUES(
	$1
	,$2
	,$3
	,$4
	,$5
)
`
	_, err = tx.Exec(sqlQ, email, token,
		now, expires, hashedPass)
	if err != nil {
		r.ins.L.Err(err, "error executing query", map[string]interface{}{
			"query": sqlQ,
		})
		if rbErr := tx.Rollback(); rbErr != nil {
			r.ins.L.Err(rbErr, "rollback failed", nil)
		}
		return "", err
	}
	return token, nil
}

// ActivateUser confirms an user email with its activation token.
func (r *RepoSQLX) ActivateUser(token string) (*User, error) {
	master := r.sqlDB.Master()
	tx, err := master.Beginx()
	if err != nil {
		return nil, err
	}

	now := time.Now()

	findQ := `
SELECT
	token
	,email
	,requested
	,expires
	,password
	,consumed
FROM
	user_registration_requests
WHERE
	token=$1
FOR UPDATE
`
	row := tx.QueryRowx(findQ, token)
	if row == nil {
		if rbErr := tx.Rollback(); rbErr != nil {
			r.ins.L.Err(rbErr, "rollback failed", nil)
		}
		return nil, ErrNotFound
	}
	rq := registrationRequest{}
	var consumed *time.Time
	if err = row.Scan(
		&rq.token,
		&rq.email,
		&rq.requested,
		&rq.expires,
		&rq.password,
		&consumed,
	); err != nil {
		r.ins.L.Err(err, fmt.Sprintf("cannot scan row %s", err.Error()), nil)
		return nil, err
	}
	if consumed != nil {
		rq.consumed = *consumed
	}

	if !rq.consumed.IsZero() {
		if rbErr := tx.Rollback(); rbErr != nil {
			r.ins.L.Err(rbErr, "rollback failed", nil)
		}
		return nil, ErrConsumed
	}

	if now.After(rq.expires) {
		if rbErr := tx.Rollback(); rbErr != nil {
			r.ins.L.Err(rbErr, "rollback failed", nil)
		}
		return nil, ErrExpired
	}

	consumeQ := `
UPDATE
	user_registration_requests
SET consumed=$2
WHERE
	token=$1
`
	if _, err := tx.Exec(consumeQ, token, now); err != nil {
		if rbErr := tx.Rollback(); rbErr != nil {
			r.ins.L.Err(rbErr, "rollback failed", nil)
		}
		return nil, err
	}

	id := ids.NewIDGenerator().MustNew()
	createUserQ := `
INSERT INTO users(
	id
	,email
	,password
	,created
)
VALUES(
	$1
	,$2
	,$3
	,$4
)
`
	if _, err := tx.Exec(createUserQ, id.ToUUID(), rq.email, rq.password, now); err != nil {
		if rbErr := tx.Rollback(); rbErr != nil {
			r.ins.L.Err(rbErr, "rollback failed", nil)
		}
		return nil, err
	}

	if err := tx.Commit(); err != nil {
		if rbErr := tx.Rollback(); rbErr != nil {
			r.ins.L.Err(rbErr, "rollback failed", nil)
		}
		r.ins.L.Err(err, "cannot complete transaction", map[string]interface{}{
			"query": createUserQ,
		})
		return nil, err
	}

	return &User{
		ID:      id,
		Email:   rq.email,
		Created: now,
	}, nil
}

// CreatePasswordResetRequest returns a token for password reset.
func (r *RepoSQLX) CreatePasswordResetRequest(email string) (*User, string, error) {
	token := r.createToken(email)

	master := r.sqlDB.Master()
	tx, err := master.Beginx()
	if err != nil {
		return nil, "", err
	}

	u := r.getUserByEmail(tx, email)
	if u == nil {
		if rbErr := tx.Rollback(); rbErr != nil {
			r.ins.L.Err(rbErr, "rollback failed", nil)
		}
		return nil, "", ErrNotFound
	}

	clearOldResetPasswordTokensQ := `
UPDATE
	user_resetpasswords
SET
	expires=requested
WHERE
	user_id=$1
	AND consumed IS NULL
`

	if _, err := tx.Exec(clearOldResetPasswordTokensQ, u.ID.ToUUID()); err != nil {
		// just log the error, we don't care much about stale old tokens
		r.ins.L.Err(err, "cannot clean existing reset tokens", nil)
	}

	expirationHours := 24
	now := time.Now()
	expires := now.Add(time.Duration(expirationHours) * time.Hour)
	insertTokenQ := `
INSERT INTO user_resetpasswords(
	user_id
	,token
	,requested
	,expires
)
VALUES (
	$1
	,$2
	,$3
	,$4
)
`
	if _, err := tx.Exec(insertTokenQ, u.ID.ToUUID(), token, now, expires); err != nil {
		if rbErr := tx.Rollback(); rbErr != nil {
			r.ins.L.Err(rbErr, "rollback failed", nil)
		}
		r.ins.L.Err(err, "cannot create reset password token", nil)
		return nil, "", fmt.Errorf("cannot create reset password token for %s: %w",
			u.ID.ToUUID(), err)
	}

	if err := tx.Commit(); err != nil {
		r.ins.L.Err(err, "cannot commit resest password request", nil)
		return nil, "", err
	}
	return u, token, nil
}

type resetPasswordRequest struct {
	userID    string
	token     string
	requested time.Time
	expires   time.Time
	consumed  time.Time
}

// ResetPassword changes the pasword for the user associated with
// the given reset password token
func (r *RepoSQLX) ResetPassword(token string, password string) (*User, error) {
	master := r.sqlDB.Master()
	tx, err := master.Beginx()
	if err != nil {
		return nil, err
	}

	checkTokenQ := `
SELECT
	user_id
	,token
	,requested
	,expires
	,consumed
FROM user_resetpasswords
WHERE
	token = $1
`
	var rpr resetPasswordRequest
	row := tx.QueryRowx(checkTokenQ, token)
	var consumed *time.Time
	if err := row.Scan(
		&rpr.userID,
		&rpr.token,
		&rpr.requested,
		&rpr.expires,
		&consumed); err != nil {
		if rbErr := tx.Rollback(); rbErr != nil {
			r.ins.L.Err(rbErr, "rollback failed", nil)
		}
		r.ins.L.Err(err, "cannot scan result", map[string]interface{}{
			"query": checkTokenQ,
		})
		return nil, err
	}
	if consumed != nil {
		rpr.consumed = *consumed
	}

	if !rpr.consumed.IsZero() {
		if rbErr := tx.Rollback(); rbErr != nil {
			r.ins.L.Err(rbErr, "rollback failed", nil)
		}
		return nil, fmt.Errorf("token already consumed")
	}

	now := time.Now()
	if now.After(rpr.expires) {
		if rbErr := tx.Rollback(); rbErr != nil {
			r.ins.L.Err(rbErr, "rollback failed", nil)
		}
		return nil, fmt.Errorf("token expired")
	}

	consumeTokenQ := `
UPDATE user_resetpasswords
SET
	consumed = $1
WHERE
	token = $2
`
	if _, err := tx.Exec(consumeTokenQ, now, token); err != nil {
		if rbErr := tx.Rollback(); rbErr != nil {
			r.ins.L.Err(rbErr, "rollback failed", nil)
		}
		werr := fmt.Errorf("cannot consume token %s: %w", token, err)
		return nil, werr
	}

	passHash := r.passwordHash(password)
	updatePasswordQ := `
UPDATE users
SET
	password = $1
WHERE
	id = $2
`

	if _, err := tx.Exec(updatePasswordQ, passHash, rpr.userID); err != nil {
		if rbErr := tx.Rollback(); rbErr != nil {
			r.ins.L.Err(rbErr, "rollback failed", nil)
		}
		werr := fmt.Errorf("cannot set password %s: %w", token, err)
		return nil, werr
	}

	var id ids.ID
	if err := id.FromUUID(rpr.userID); err != nil {
		r.ins.L.Err(err, "bad userID format", nil)
		return nil, err
	}
	u := r.getUserByID(tx, id)
	if u == nil {
		if rbErr := tx.Rollback(); rbErr != nil {
			r.ins.L.Err(rbErr, "rollback failed", nil)
		}
		err := fmt.Errorf("cannot get user %s", rpr.userID)
		return nil, err
	}
	if err := tx.Commit(); err != nil {
		return nil, err
	}
	return u, nil
}

// CheckPassword return the user ID for a user from its email
// and password.
func (r *RepoSQLX) CheckPassword(email string, password string) (ids.ID, error) {
	master := r.sqlDB.Master()
	var userID ids.ID
	getHashedPwdQ := `
SELECT
	id
	,password
FROM users
WHERE
	email = $1
`
	row := master.QueryRowx(getHashedPwdQ, email)
	var strID string
	var hashedPwd string
	if err := row.Scan(&strID, &hashedPwd); err != nil {
		return userID, err
	}
	if !r.checkPassword(password, hashedPwd) {
		return userID, fmt.Errorf("wrong password")
	}
	err := userID.FromUUID(strID)
	return userID, err
}

// DeleteUser hard deletes a user (and all its related
// data will be deleted too)
func (r *RepoSQLX) DeleteUser(email string) error {
	master := r.sqlDB.Master()
	deleteQ := `
DELETE FROM users
WHERE
    email = $1
`
	_, err := master.Exec(deleteQ, email)
	return err
	// TODO: depending on the database we could check the RowsAffected
	// if rows.RowsAffected() == 0, we could return a not found
}

// ListUsers lists users with pagination
func (r *RepoSQLX) ListUsers(from ids.ID, limit int, backwards bool) ([]User, error) {
	// TODO: check if we should do an union with the `user_registration_requests` to
	// also list those users that have not activated the account
	master := r.sqlDB.Master()
	var rows *sqlx.Rows
	var err error

	if from.IsZero() {
		q := `
SELECT
    id
    , email
    , created
FROM users
LIMIT $1
`
		rows, err = master.Queryx(q, limit)
		if err != nil {
			return []User{}, err
		}
	} else {
		if !backwards {
			q := `
SELECT
    id
    , email
    , created
FROM users
WHERE 
    id > $1
ORDER BY id
LIMIT $2
`
			rows, err = master.Queryx(q, from.ToUUID(), limit)
			if err != nil {
				return []User{}, err
			}
		} else {
			q := `
WITH backpage AS (
    SELECT 
        id 
    FROM users
    WHERE id < $1
    ORDER BY id DESC
    LIMIT $2
)
SELECT
    id
    , email
    , created
FROM users
WHERE id IN (SELECT id FROM backpage)
ORDER BY id
`
			rows, err = master.Queryx(q, from.ToUUID(), limit)
			if err != nil {
				return []User{}, err
			}
		}
	}

	results := make([]User, 0, limit)
	var u User
	for rows.Next() {
		if err := rows.StructScan(&u); err != nil {
			return []User{}, err
		}
		results = append(results, u)
	}
	return results, nil
}
