package users

import (
	"fmt"
	"time"

	"github.com/dhontecillas/hfw/pkg/ids"
)

// User represents a user in the system
type User struct {
	ID      ids.ID
	Email   string
	Created time.Time
}

func (u *User) String() string {
	return fmt.Sprintf("%s (%d) -> %s", u.ID.ToUUID(), u.Created.UnixNano(), u.Email)
}
