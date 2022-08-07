package users

import (
	"testing"
	"time"

	hfwtest "github.com/dhontecillas/hfw/testing"

	"github.com/dhontecillas/hfw/pkg/ids"
)

func Test_RepoSQLX_HappyPath(t *testing.T) {
	email := "foo@example.com"
	pass := "bar"

	deps := hfwtest.BuildExternalServices()

	r := NewRepoSQLX(deps.Insighter(), deps.SQL, "tokenSalt")

	token, err := r.CreateInactiveUser(email, pass)
	if err != nil {
		t.Errorf("cannot create inactive user: %s", err.Error())
		return
	}

	if len(token) == 0 {
		t.Errorf("expected activation token")
		return
	}

	u, err := r.ActivateUser(token)
	if err != nil {
		t.Errorf("cannot activate user: %s", err)
		return
	}

	if u.Email != email {
		t.Errorf("email, want: %s, got: %s", email, u.Email)
		return
	}

	var emptyID ids.ID
	if u.ID == emptyID {
		t.Errorf("empty user id")
		return
	}

	var zeroTime time.Time
	if u.Created == zeroTime {
		t.Errorf("empty created time")
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
		t.Errorf("expected")
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
}
