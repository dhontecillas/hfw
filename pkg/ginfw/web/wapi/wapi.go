package wapi

import (
	"net/http"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	// csrf "github.com/utrack/gin-csrf"
)

// Middleware sets up a restrictive CORS Middleware and
// some extra headers
func Middleware() gin.HandlerFunc {
	corsConf := cors.DefaultConfig()
	corsConf.AllowAllOrigins = false
	corsConf.AllowOrigins = []string{
		"http://localhost:9999",
	}
	corsConf.AllowMethods = []string{
		"GET", "POST", "PUT", "PATCH", "DELETE", "HEAD", // "OPTIONS",
	}
	corsConf.AllowHeaders = append(corsConf.AllowHeaders, "X-Csrf-Token")
	corsConf.AllowHeaders = append(corsConf.AllowHeaders, "X-Client")
	corsConf.AllowHeaders = append(corsConf.AllowHeaders, "Origin")
	corsFn := cors.New(corsConf)
	return func(c *gin.Context) {
		c.Header("X-Frame-Options", "DENY")
		c.Header("X-XSS-Protection", "1, mode=block")
		c.Header("Content-Security-Policy", "default-src 'self'")
		c.Header("Cache-Control", "no-cache, no-store")
		corsFn(c)
	}
}

// DevMiddleware sets up a permisive CORS Middleware to
// allow frontend apis served in different port to access the web apis.
func DevMiddleware() gin.HandlerFunc {
	corsConf := cors.DefaultConfig()
	corsConf.AllowAllOrigins = false
	corsConf.AllowOrigins = []string{}
	corsConf.AllowMethods = []string{
		"GET", "POST", "PUT", "PATCH", "DELETE", "HEAD", // "OPTIONS",
	}
	corsConf.AllowCredentials = true
	corsConf.AllowHeaders = append(corsConf.AllowHeaders, "X-Csrf-Token")
	corsConf.AllowHeaders = append(corsConf.AllowHeaders, "X-Client")
	corsConf.AllowHeaders = append(corsConf.AllowHeaders, "Origin")
	return cors.New(corsConf)
}

// Use makes use of the WAPIMiddleware
func Use(r gin.IRouter, devMode bool) {
	r.Use(Middleware())
}

// GetCSRFToken returns a payload of the csrf_token
func GetCSRFToken(c *gin.Context) {
	var msg struct {
		CsrfToken string `json:"csrf_token"`
	}
	msg.CsrfToken = "tkn" // csrf.GetToken(c)
	c.JSON(http.StatusOK, msg)
}
