package users

import (
	"time"

	"github.com/dhontecillas/hfw/pkg/ids"
)

// User represents a user in the system
type User struct {
	ID      ids.ID
	Email   string
	Created time.Time
}
