package tokenapi

import (
	"testing"
	"time"

	"github.com/dhontecillas/hfw/pkg/extdeps"
	"github.com/dhontecillas/hfw/pkg/ids"
	"github.com/dhontecillas/hfw/pkg/usecases/users"
	hfwtest "github.com/dhontecillas/hfw/testing"
)

func createTestUser(deps *extdeps.ExternalServices) (*users.User, error) {
	email := "example@example.com"
	pass := "bar"
	r := users.NewRepoSQLX(deps.Insighter(), deps.SQL, "tokenSalt")
	token, err := r.CreateInactiveUser(email, pass)
	if err != nil {
		return nil, err
	}
	_, err = r.ActivateUser(token)
	if err != nil {
		return nil, err
	}
	u := r.GetUserByEmail(email)
	return u, nil
}

func Test_RepoSQLX_HappyPath(t *testing.T) {

	deps := hfwtest.BuildExternalServices()
	u, err := createTestUser(deps)
	if err != nil {
		t.Errorf("err creating user : %s", err)
	}

	r := NewRepoSQLX(deps.Insighter(), deps.SQL)

	idGen := ids.NewIDGenerator()

	key, _ := idGen.New()
	tk, err := r.CreateKey(key, u.ID, "Test Key", time.Now())
	if err != nil {
		t.Errorf("error creating key: %s", err)
		return
	}
	if key != tk.Key {
		t.Errorf("key want %s, got %s", key, tk.Key)
		return
	}

	apiKeys, err := r.ListKeys(u.ID)
	if err != nil {
		t.Errorf("error listing keys: %s for %s", err, u.ID.ToUUID())
		return
	}
	if len(apiKeys) != 1 {
		t.Errorf("want 1 apiKey, got %d", len(apiKeys))
		return
	}

	lk := apiKeys[0]
	if lk.Key.ToUUID() != tk.Key.ToUUID() || lk.Description != tk.Description {
		t.Errorf("key listing mismatch, want %#v, got %#v",
			tk, lk)
		return
	}

	getK, err := r.GetKey(tk.Key)
	if err != nil {
		t.Errorf("get key failed: %s", err)
		return
	}

	if getK == nil {
		t.Errorf("expected key, got nil")
		return
	}

	err = r.DeleteUserKey(u.ID, tk.Key)
	if err != nil {
		t.Errorf("err deleting key: %s", err)
		return
	}

	lks, err := r.ListKeys(u.ID)
	if err != nil {
		t.Errorf("err listing keys after deletion: %s", err)
		return
	}

	if len(lks) != 0 {
		t.Errorf("expected empty list for user, got %d", len(lks))
		return
	}

	_, err = r.GetKey(tk.Key)
	if err == nil {
		t.Errorf("exepected not found error")
		return
	}

}
