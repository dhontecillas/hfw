package auth

import (
	"github.com/gin-gonic/gin"

	"github.com/dhontecillas/hfw/pkg/ids"
)

const (
	userIDKey string = "HFW_UserID"
)

// GetUserID returns the current user id from a request context.
// if it exists.
func GetUserID(c *gin.Context) *ids.ID {
	v, exists := c.Get(userIDKey)
	if exists {
		if id, ok := v.(ids.ID); ok {
			return &id
		}
	}
	return nil
}

// SetUserID sets the current use id into the gin context.
func SetUserID(c *gin.Context, userID ids.ID) {
	c.Set(userIDKey, userID)
}
