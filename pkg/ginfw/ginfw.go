package ginfw

import (
	"github.com/dhontecillas/hfw/pkg/extdeps"
	"github.com/gin-gonic/gin"
)

const extServicesKey = "HFW_ExtServices"

// ExtServices returns the ExtServices instance stored
// in the gin.Context
func ExtServices(c *gin.Context) *extdeps.ExtServices {
	es, ok := c.Keys[extServicesKey].(*extdeps.ExtServices)
	if !ok {
		panic("external services not set")
	}
	return es
}

// ExtServicesMiddleware creates a new ExternalServices (or external dependencies)
// for the incoming request, and stores it in the gin context
func ExtServicesMiddleware(es *extdeps.ExternalServices) gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Set(extServicesKey, es.ExtServices())
		c.Next()
	}
}
