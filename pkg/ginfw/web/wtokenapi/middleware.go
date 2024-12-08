package wtokenapi

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/dhontecillas/hfw/pkg/extdeps"
	"github.com/dhontecillas/hfw/pkg/ids"
	"github.com/dhontecillas/hfw/pkg/tokenapi"

	"github.com/dhontecillas/hfw/pkg/ginfw/auth"
)

const (
	apiKeyHeaderName = "X-Api-Key"
)

// RequireAPIToken checks a valid token, and stores the
// userID for the token to the context
func RequireAPIToken(extDeps *extdeps.ExternalServicesBuilder) gin.HandlerFunc {
	ins := extDeps.Insighter()
	// here we construct the repository to check the api keys
	tokenAPIRepo := tokenapi.NewRepoSQLX(ins, extDeps.SQL)
	tokenAPI := tokenapi.NewTokenAPI(ins, tokenAPIRepo)

	return func(c *gin.Context) {
		strAPIKey, ok := c.Request.Header[apiKeyHeaderName]
		if !ok || len(strAPIKey) != 1 {
			c.JSON(http.StatusUnauthorized, nil)
			c.Abort()
			return
		}
		var apiKey ids.ID
		if err := apiKey.FromShuffled(strAPIKey[0]); err != nil {
			c.JSON(http.StatusUnauthorized, nil)
			c.Abort()
			return
		}

		tk, err := tokenAPI.GetKey(apiKey)
		if err != nil || tk == nil {
			c.JSON(http.StatusUnauthorized, nil)
			c.Abort()
			return
		}

		if tk.Deleted != nil {
			c.JSON(http.StatusUnauthorized, nil)
			c.Abort()
			return
		}

		auth.SetUserID(c, tk.UserID)
	}
}
