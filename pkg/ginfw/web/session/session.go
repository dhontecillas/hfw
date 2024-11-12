package session

import (
	"fmt"
	"html/template"
	"net/http"

	"github.com/gin-contrib/sessions"
	"github.com/gin-contrib/sessions/redis"
	"github.com/gin-gonic/gin"
	csrf "github.com/utrack/gin-csrf"

	"github.com/dhontecillas/hfw/pkg/ginfw/auth"
	"github.com/dhontecillas/hfw/pkg/ids"
)

const keySessionName string = "sess"
const keyUserID string = "strUserID"

// LogInSession is a redis store for user sessions
var LogInSession redis.Store

// RedisConf contains the configuration to
// maintain user session backed by a Redis server
type RedisConf struct {
	MaxIdleConnections int
	Host               string
	Password           string
	SecretKeyPair      string
}

// Conf maintains the session configuration
type Conf struct {
	RedisConf  RedisConf
	CsrfSecret string
	IsDevelop  bool
}

// Setup initializes the global `LogInSession` redis.Store instance
func Setup(conf *Conf) error {
	var err error
	LogInSession, err = redis.NewStore(conf.RedisConf.MaxIdleConnections, "tcp",
		conf.RedisConf.Host, conf.RedisConf.Password,
		[]byte(conf.RedisConf.SecretKeyPair))
	return err
}

// Use adds the session and csrf token middleware
func Use(r gin.IRoutes, conf *Conf) {
	err := Setup(conf)
	if err != nil {
		panic(fmt.Sprintf("cannot set up session\n%#v\n: %s", *conf, err.Error()))
	}
	r.Use(sessions.Sessions(keySessionName, LogInSession))
	r.Use(csrf.Middleware(csrf.Options{
		Secret: conf.CsrfSecret,
		ErrorFunc: func(c *gin.Context) {
			c.String(400, "Bad CSRF token")
			c.Abort()
		},
	}))
}

// GetUserID returns the user ID of an authenticated request
func GetUserID(c *gin.Context) string {
	s := sessions.Default(c)
	v := s.Get(keyUserID)
	userID, ok := v.(string)
	if !ok {
		return ""
	}
	return userID
}

// ClearUserID remove a user session (effectively logging out the user)
func ClearUserID(c *gin.Context) {
	s := sessions.Default(c)
	s.Clear()
	// TODO: check if we should return an error
	_ = s.Save()
}

// SetUserID sets a userId to a session (effectivly logging in the user)
func SetUserID(c *gin.Context, userID string) {
	s := sessions.Default(c)
	s.Set(keyUserID, userID)
	// TODO: check if we should return an error
	_ = s.Save()
}

// GetCSRFTokenInput returns a hidden input tag with the token
func GetCSRFTokenInput(c *gin.Context) template.HTML {
	token := csrf.GetToken(c)
	inputTag := fmt.Sprintf("<input type=\"hidden\" name=\"_csrf\" value=\"%s\">", token)
	return template.HTML(inputTag)
}

// AuthRequired is a middleware to check there is a logged in user
// in the current session
func AuthRequired() gin.HandlerFunc {
	return func(c *gin.Context) {
		strUserID := GetUserID(c)
		if len(strUserID) == 0 {
			c.JSON(http.StatusUnauthorized, nil)
			c.Abort()
			return
		}
		var userID ids.ID
		if err := userID.FromUUID(strUserID); err != nil {
			c.JSON(http.StatusServiceUnavailable, nil)
			c.Abort()
			return
		}
		auth.SetUserID(c, userID)
	}
}
