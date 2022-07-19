package tokenapi

import (
	"time"

	"github.com/dhontecillas/hfw/pkg/db"
	"github.com/dhontecillas/hfw/pkg/ids"
	"github.com/dhontecillas/hfw/pkg/obs"
)

// RepoSQLX implments the token api repository with sqlx
type RepoSQLX struct {
	sqlDB db.SQLDB
	ins   *obs.Insighter
}

// NewRepoSQLX creates a new RepoSQLX
func NewRepoSQLX(ins *obs.Insighter, sqlDB db.SQLDB) *RepoSQLX {
	return &RepoSQLX{
		sqlDB: sqlDB,
		ins:   ins,
	}
}

type sqlxTokenAPIKey struct {
	ID          string
	UserID      string
	Created     time.Time
	Deleted     *time.Time
	LastUsed    *time.Time
	Description string
}

func (st *sqlxTokenAPIKey) fromSQLX(t *APIKey) error {
	if err := t.Key.FromUUID(st.ID); err != nil {
		return err
	}
	if err := t.UserID.FromUUID(st.UserID); err != nil {
		return err
	}
	t.Created = st.Created
	t.Deleted = st.Deleted
	t.LastUsed = st.LastUsed
	t.Description = st.Description
	return nil
}

/*
TODO: remove this ? we are not using it right now because
the creation is made with some parameters.
func (st *sqlxTokenAPIKey) toSQLX(t *APIKey) error {
	st.ID = t.Key.ToUUID()
	st.UserID = t.UserID.ToUUID()
	st.Created = t.Created
	st.Deleted = t.Deleted
	st.LastUsed = t.LastUsed
	st.Description = t.Description
	return nil
}
*/

// CreateKey creates a new api key.
func (r *RepoSQLX) CreateKey(key ids.ID, userID ids.ID, description string,
	created time.Time) (*APIKey, error) {

	sqlQ := `
INSERT INTO tokenapi_keys(
	id
	,user_id
	,created
	,deleted
	,last_used
	,description
)
VALUES(
	$1
	,$2
	,$3
	,NULL
	,NULL
	,$4
)
`
	strKey := key.ToUUID()
	strUserID := userID.ToUUID()

	master := r.sqlDB.Master()
	// TODO: this does not need a transaction, this can be done
	// in a single request to the DB
	tx, err := master.Beginx()
	if err != nil {
		return nil, err
	}

	_, err = tx.Exec(sqlQ, strKey, strUserID, created, description)
	if err != nil {
		_ = tx.Rollback()
		// TODO: if rollback failed .. what?
		return nil, err
	}
	err = tx.Commit()
	if err != nil {
		return nil, err
	}

	return &APIKey{
		Key:         key,
		UserID:      userID,
		Created:     created,
		Description: description,
	}, nil
}

// GetKey retrieves an existing api key by id.
func (r *RepoSQLX) GetKey(key ids.ID) (*APIKey, error) {
	sqlQ := `
SELECT
	id AS ID
	,user_id AS UserID
	,created AS Created
	,deleted AS Deleted
	,last_used AS LastUsed
	,description AS Description
FROM tokenapi_keys
WHERE
	id = $1
`
	strKey := key.ToUUID()
	master := r.sqlDB.Master()
	row := master.QueryRowx(sqlQ, strKey)
	if row == nil {
		return nil, ErrNotFound
	}
	var sqlT sqlxTokenAPIKey
	if err := row.StructScan(&sqlT); err != nil {
		return nil, err
	}
	tk := APIKey{}
	if err := sqlT.fromSQLX(&tk); err != nil {
		return nil, err
	}
	return &tk, nil
}

// ListKeys returns a full list of api keys for a user.
func (r *RepoSQLX) ListKeys(userID ids.ID) ([]APIKey, error) {
	sqlQ := `
SELECT
	id AS ID
	,user_id AS UserID
	,created AS Created
	,deleted AS Deleted
	,last_used AS LastUsed
	,description AS Description
FROM tokenapi_keys
WHERE
	user_id = $1
`
	strUserID := userID.ToUUID()
	master := r.sqlDB.Master()
	rows, err := master.Queryx(sqlQ, strUserID)
	if err != nil {
		return nil, ErrNotFound
	}
	defer rows.Close()

	tks := make([]APIKey, 0, 16)
	var sqlT sqlxTokenAPIKey
	var tk APIKey
	for rows.Next() {
		if err := rows.StructScan(&sqlT); err != nil {
			return nil, err
		}
		if err := sqlT.fromSQLX(&tk); err != nil {
			return nil, err
		}
		tks = append(tks, tk)
	}
	return tks, nil
}

// DeleteKey deletes an existing api key by id.
func (r *RepoSQLX) DeleteKey(key ids.ID) error {
	// TODO
	return nil
}

// DeleteUserKey deletes an existing api key by id checking
// that it belongs to the given user.
func (r *RepoSQLX) DeleteUserKey(userID ids.ID, key ids.ID) error {
	sqlQ := `
DELETE FROM tokenapi_keys
WHERE
	id = $1
	AND user_id = $2
`
	strUserID := userID.ToUUID()
	strKeyID := key.ToUUID()

	master := r.sqlDB.Master()
	_, err := master.Exec(sqlQ, strKeyID, strUserID)
	if err != nil {
		return err
	}
	return nil
}
