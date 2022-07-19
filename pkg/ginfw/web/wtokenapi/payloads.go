package wtokenapi

import (
	"time"

	"github.com/dhontecillas/hfw/pkg/tokenapi"
)

// CreatePayload is the required payload to create an API token.
type CreatePayload struct {
	Description string `json:"description" binding:"required"`
}

// DeletePayload is the required payload to delete an API token.
type DeletePayload struct {
	Key string `json:"key" binding:"required"`
}

// TokenAPIKey has the data for an API token
type TokenAPIKey struct {
	Key         string `json:"key"`
	Created     string `json:"created"`
	Description string `json:"description"`
}

// TokenAPIKeyList has a list of TokenAPIKey's
type TokenAPIKeyList struct {
	APIKeys []TokenAPIKey `json:"apiKeys"`
}

func fromTokenAPI(t *tokenapi.APIKey) *TokenAPIKey {
	return &TokenAPIKey{
		Key:         t.Key.ToShuffled(),
		Created:     t.Created.Format(time.RFC3339),
		Description: t.Description,
	}
}

func fromTokenAPISlice(t []tokenapi.APIKey) TokenAPIKeyList {
	res := TokenAPIKeyList{
		APIKeys: make([]TokenAPIKey, 0, len(t)),
	}
	for _, tak := range t {
		res.APIKeys = append(res.APIKeys, *fromTokenAPI(&tak))
	}
	return res
}
