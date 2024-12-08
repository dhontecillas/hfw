package wusers

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/dhontecillas/hfw/pkg/ginfw"
	"github.com/dhontecillas/hfw/pkg/ginfw/web/session"
)

const (
	// PathIsLoggedIn is the path to check if a user is logged in.
	PathIsLoggedIn string = "loggedin"
)

// WAPIRoutes create the routes for the endpoints to be used from the
// web app.
func WAPIRoutes(r gin.IRouter, actionPaths ActionPaths) {
	r.POST(PathRegister,
		emailRegistrationMiddleware(WAPIRegister, actionPaths))
	r.POST(PathLogin,
		emailRegistrationMiddleware(WAPILogin, actionPaths))
	r.POST(PathRequestPasswordReset,
		emailRegistrationMiddleware(WAPIRequestPasswordReset, actionPaths))
	r.POST(PathLogout, WAPILogout)
	r.GET(PathIsLoggedIn, WAPIIsLoggedIn)
	r.POST(PathResetPassword,
		emailRegistrationMiddleware(WAPIResetPasswordWithToken, actionPaths))
}

// OKRes has the result for a successful operation.
type OKRes struct {
	Success bool `json:"success"`
}

// FailRes has the result for a failed operation.
type FailRes struct {
	Success bool   `json:"success"`
	Error   string `json:"error"`
}

// WAPIRegister is the handler for the register user endpoint.
func WAPIRegister(c *gin.Context, actionPaths *ActionPaths) {
	// for API calls we do not check repeated password, as we
	// leave that for the UI
	p := LoginPayload{}
	err := c.ShouldBindJSON(&p)
	if err != nil {
		deps := ginfw.ExtServices(c)
		deps.Ins.L.Err(err, "cannot bind the payload", nil)
		c.JSON(http.StatusBadRequest, FailRes{Error: err.Error()})
		return
	}
	regUC := buildController(c, actionPaths)
	err = regUC.Register(p.Email, p.Password)
	if err != nil {
		// we do not leak the reason why the registration failed
		c.JSON(http.StatusOK, OKRes{Success: true})
		return
	}
	c.JSON(http.StatusOK, OKRes{Success: true})
}

// WAPILogin is the handler for the login user endpoint.
func WAPILogin(c *gin.Context, actionPaths *ActionPaths) {
	p := LoginPayload{}
	err := c.ShouldBindJSON(&p)
	if err != nil {
		c.JSON(http.StatusBadRequest, FailRes{Error: err.Error()})
		return
	}
	regUC := buildController(c, actionPaths)
	userID, err := regUC.Login(p.Email, p.Password)
	if err != nil {
		c.JSON(http.StatusBadRequest, FailRes{Error: err.Error()})
		return
	}
	session.SetUserID(c, userID.ToUUID())
	c.JSON(http.StatusOK, OKRes{Success: true})
}

// WAPIRequestPasswordReset is the handler for the request password
// reset endpoint.
func WAPIRequestPasswordReset(c *gin.Context, actionPaths *ActionPaths) {
	p := RequestResetPasswordPayload{}
	err := c.ShouldBindJSON(&p)
	if err != nil {
		// c.JSON(http.StatusBadRequest, FailRes{Error: err.Error()})
		// we alway return Ok, to not leak if an email exists or not
		deps := ginfw.ExtServices(c)
		deps.Ins.L.Err(err, "cannot bind the payload", nil)
		c.JSON(http.StatusOK, OKRes{Success: true})
		return
	}
	regUC := buildController(c, actionPaths)
	err = regUC.RequestResetPassword(p.Email)
	if err != nil {
		// we do not leak info about if the user exists or not
		deps := ginfw.ExtServices(c)
		deps.Ins.L.Err(err, "cannot request password reset", nil)
		c.JSON(http.StatusOK, OKRes{Success: true})
	}
	c.JSON(http.StatusOK, OKRes{Success: true})
}

// WAPIResetPasswordWithToken is the handler to change a password with
// a reset password token.
func WAPIResetPasswordWithToken(c *gin.Context, actionPaths *ActionPaths) {
	p := ResetPasswordWithTokenPayload{}
	err := c.ShouldBindJSON(&p)
	if err != nil {
		deps := ginfw.ExtServices(c)
		deps.Ins.L.Err(err, "cannot bind the payload", nil)
		// c.JSON(http.StatusBadRequest, FailRes{Error: err.Error()})
		// we alway return Ok, to not leak if an email exists or not
		c.JSON(http.StatusOK, OKRes{Success: true})
		return
	}
	regUC := buildController(c, actionPaths)
	err = regUC.ResetPasswordWithToken(p.Token, p.Password)
	if err != nil {
		// we do not leak info about if the user exists or not
		deps := ginfw.ExtServices(c)
		deps.Ins.L.Err(err, "cannot reset password with token", nil)
		// c.JSON(http.StatusBadRequest, FailRes{Error: err.Error()})
		// we alway return Ok, to not leak if an email exists or not
		c.JSON(http.StatusOK, OKRes{Success: true})
	}
	c.JSON(http.StatusOK, OKRes{Success: true})
}

// WAPILogout is the handler to log out a user.
func WAPILogout(c *gin.Context) {
	session.ClearUserID(c)
	c.JSON(http.StatusOK, OKRes{Success: true})
}

// WAPIIsLoggedIn is the handler to check if a user
// is logged in.
func WAPIIsLoggedIn(c *gin.Context) {
	if session.GetUserID(c) == "" {
		c.JSON(http.StatusNotFound, OKRes{Success: false})
		return
	}
	c.JSON(http.StatusOK, OKRes{Success: true})
}
