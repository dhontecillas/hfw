package tokenapi

import (
	"time"

	"github.com/dhontecillas/hfw/pkg/ids"
	"github.com/dhontecillas/hfw/pkg/obs"
)

// Repo defines the storage contract for the TokenAPI functionality
type Repo interface {
	// CreateKey creates a new api key.
	CreateKey(key ids.ID, userID ids.ID, description string,
		created time.Time) (*APIKey, error)
	// GetKey retrieves an existing api key by id.
	GetKey(key ids.ID) (*APIKey, error)
	// ListKeys returns a full list of api keys for a user.
	ListKeys(userID ids.ID) ([]APIKey, error)
	// DeleteKey deletes an existing api key by id.
	DeleteKey(key ids.ID) error
	// DeleteUserKey deletes an existing api key by id checking
	// that it belongs to the given user.
	DeleteUserKey(userID ids.ID, key ids.ID) error
}

// TokenAPI defines the interface to interact with api tokens.
type TokenAPI interface {
	CreateKey(userID ids.ID, description string) (*APIKey, error)
	DeleteKey(userID ids.ID, key ids.ID) error
	GetKey(key ids.ID) (*APIKey, error)
	ListKeys(userID ids.ID, onlyActive bool) ([]APIKey, error)
}

type tokenAPI struct {
	ins  *obs.Insighter
	repo Repo
}

// NewTokenAPI creates a new TokenAPI instance
func NewTokenAPI(ins *obs.Insighter, repo Repo) TokenAPI {
	return &tokenAPI{
		ins:  ins,
		repo: repo,
	}
}

func (t *tokenAPI) CreateKey(userID ids.ID, description string) (*APIKey, error) {
	idGen := ids.NewIDGenerator()
	key, err := idGen.New()
	if err != nil {
		t.ins.L.Err(err, "cannot create unique id")
		return nil, err
	}
	t.ins.L.WarnMsg("create key").Str("key", key.ToUUID()).
		Str("userID", userID.ToUUID()).Send()
	res, err := t.repo.CreateKey(key, userID, description, time.Now())
	if err != nil {
		t.ins.L.Err(err, "cannot create key")
	}
	return res, err
}

func (t *tokenAPI) ListKeys(userID ids.ID, onlyActive bool) ([]APIKey, error) {
	keys, err := t.repo.ListKeys(userID)
	if err != nil {
		t.ins.L.Err(err, "cannot list user keys")
		return nil, err
	}
	return keys, err
}

func (t *tokenAPI) DeleteKey(userID ids.ID, key ids.ID) error {
	err := t.repo.DeleteUserKey(userID, key)
	if err != nil {
		t.ins.L.Err(err, "cannot delete key")
		return err
	}
	return nil
}

func (t *tokenAPI) GetKey(key ids.ID) (*APIKey, error) {
	res, err := t.repo.GetKey(key)
	if err != nil {
		t.ins.L.Err(err, "cannot get key")
		return nil, err
	}
	return res, nil
}
