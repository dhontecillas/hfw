package users

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"testing"
	"time"

	hfwtest "github.com/dhontecillas/hfw/testing"

	"github.com/dhontecillas/hfw/pkg/ids"
)

func Test_RepoSQLX_HappyPath(t *testing.T) {
	email, pass := hfwtest.RandomEmailAndPassword()

	deps := hfwtest.BuildExternalServices()

	r := NewRepoSQLX(deps.Insighter(), deps.SQL, "tokenSalt")

	u := r.GetUserByEmail(email)
	if u != nil {
		t.Errorf("the user should not exist in the database")
		return
	}

	var emptyID ids.ID
	u = r.GetUserByID(emptyID)
	if u != nil {
		t.Errorf("the user with empty id should not exist")
		return
	}

	token, err := r.CreateInactiveUser(email, pass)
	if err != nil {
		t.Errorf("cannot create inactive user: %s", err.Error())
		return
	}

	if len(token) == 0 {
		t.Errorf("expected activation token")
		return
	}

	u, err = r.ActivateUser(token)
	if err != nil {
		t.Errorf("cannot activate user: %s", err)
		return
	}

	if u.Email != email {
		t.Errorf("email, want: %s, got: %s", email, u.Email)
		return
	}

	if u.ID == emptyID {
		t.Errorf("empty user id")
		return
	}

	uByID := r.GetUserByID(u.ID)
	if uByID == nil || uByID.ID != u.ID {
		t.Errorf("existing user cannot be fetched by ID")
		return
	}

	var zeroTime time.Time
	if u.Created == zeroTime {
		t.Errorf("empty created time")
		return
	}

	// check we cannot create a user with the same email:
	_, err = r.CreateInactiveUser(email, pass)
	if err == nil {
		t.Errorf("should not be able to create a user with the same email")
		return
	}

	_, _, err = r.CreatePasswordResetRequest("baduser@example.com")
	if err == nil {
		t.Errorf("non existing user should return an error")
		return
	}

	pwdUser, pwdToken, err := r.CreatePasswordResetRequest(email)
	if err != nil {
		t.Errorf("cannot create password reset request: %s", err)
		return
	}

	if pwdUser == nil {
		t.Errorf("expected non nil user")
		return
	}
	if len(pwdToken) == 0 {
		t.Errorf("expected a password token for completing a reset password")
		return
	}

	// a random token does not work for resetting the password:
	badTkn := sha256.Sum256([]byte("randombadtoken"))
	_, err = r.ResetPassword(hex.EncodeToString(badTkn[:]), "foo")
	if err == nil {
		t.Errorf("should not find a user with a random bad token")
		return
	}

	resetUser, err := r.ResetPassword(pwdToken, "baz")
	if err != nil {
		t.Errorf("cannot reset password: %s token: %s", err.Error(), pwdToken)
		return
	}

	if resetUser == nil {
		t.Errorf("unexpected nil resetUser")
		return
	}

	_, err = r.CheckPassword(email, "baz")
	if err != nil {
		t.Errorf("cannot login with new password: %s", err.Error())
		return
	}

	err = r.DeleteUser(email)
	if err != nil {
		t.Errorf("cannot delete user: %s", err.Error())
		return
	}

	_, err = r.CheckPassword(email, "baz")
	if err == nil {
		t.Errorf("user should not exist: %s", email)
	}

}

func Test_RepoSQLX_ListUsers(t *testing.T) {
	deps := hfwtest.BuildExternalServices()
	r := NewRepoSQLX(deps.Insighter(), deps.SQL, "tokenSalt")
	// create a list of users
	for i := 0; i < 100; i++ {
		email := fmt.Sprintf("ul_%d@example.com", i)
		pass := fmt.Sprintf("ul_%d", i)
		token, _ := r.CreateInactiveUser(email, pass)
		_, _ = r.ActivateUser(token)
	}

	var id ids.ID // the zero ID
	users, e := r.ListUsers(id, 10, false)
	if e != nil {
		t.Errorf("cannot list users page 0: %s", e.Error())
		return
	}
	if len(users) != 10 {
		t.Errorf("page 0: want 10 users, got %d", len(users))
		return
	}
	fmt.Printf("Page 0")
	for idx, u := range users {
		fmt.Printf("%d :: %s\n", idx, u.String())
	}

	p0u0 := users[0]
	p0u9 := users[9]

	users, e = r.ListUsers(p0u9.ID, 10, false)
	if e != nil {
		t.Errorf("cannot list users page 1: %s", e.Error())
		return
	}
	if len(users) != 10 {
		t.Errorf("page 1: want 10 users, got %d", len(users))
		return
	}
	fmt.Printf("Page 1")
	for idx, u := range users {
		fmt.Printf("%d :: %s\n", idx, u.String())
	}

	p1u0 := users[0]
	p1u9 := users[1]

	if p1u0.ID == p0u0.ID {
		t.Errorf("bad pagination p0u0 %s == p1u0 %s", p0u0.ID, p1u0.ID)
		return
	}
	if p1u9.ID == p0u9.ID {
		t.Errorf("bad pagination p0u0 %s == p1u0 %s", p0u9.ID, p1u9.ID)
		return
	}

	users, e = r.ListUsers(p1u0.ID, 10, true)
	if e != nil {
		t.Errorf("cannot list users backward page 0: %s", e.Error())
		return
	}
	if len(users) != 10 {
		t.Errorf("backward page 0: want 10 users, got %d", len(users))
		for idx, u := range users {
			t.Errorf("  ->  %d -> %s", idx, u.ID.ToUUID())
		}
		t.Errorf("cursor id -> %s", p1u0.String())
		return
	}

	if users[0].ID != p0u0.ID {
		t.Errorf("backward page 0, user 0 different from page 0, user 0")
		return
	}

}
