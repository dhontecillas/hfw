package tokenapi

import (
	"time"

	"github.com/dhontecillas/hfw/pkg/ids"
)

// APIKey contains the data regarding an API token key.
type APIKey struct {
	Key         ids.ID
	UserID      ids.ID
	Created     time.Time
	Deleted     *time.Time
	LastUsed    *time.Time
	Description string
}
