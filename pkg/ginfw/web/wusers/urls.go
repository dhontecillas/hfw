package wusers

import (
	"errors"
	"fmt"
	"net/http"
	"path/filepath"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/gin-gonic/gin/binding"

	"github.com/dhontecillas/hfw/pkg/ginfw"
	"github.com/dhontecillas/hfw/pkg/ginfw/web/session"
	"github.com/dhontecillas/hfw/pkg/usecases/users"
)

// Routes setup the routes for the user registration, login
// and reset password flows.
func Routes(r gin.IRouter, actionPaths ActionPaths) {
	r.POST(PathRegister,
		emailRegistrationMiddleware(Register, actionPaths))
	r.GET(PathRegister,
		emailRegistrationMiddleware(RegisterForm, actionPaths))
	r.GET(PathActivate,
		emailRegistrationMiddleware(Activate, actionPaths))
	r.POST(PathLogin,
		emailRegistrationMiddleware(Login, actionPaths))
	r.GET(PathLogin,
		emailRegistrationMiddleware(LoginForm, actionPaths))
	r.POST(PathLogout,
		emailRegistrationMiddleware(Logout, actionPaths))
	r.GET(PathRequestPasswordReset,
		emailRegistrationMiddleware(RequestResetPasswordForm, actionPaths))
	r.POST(PathRequestPasswordReset,
		emailRegistrationMiddleware(RequestResetPassword, actionPaths))
	r.GET(PathResetPassword,
		emailRegistrationMiddleware(ResetPasswordForm, actionPaths))
	r.POST(PathResetPassword,
		emailRegistrationMiddleware(ResetPasswordWithToken, actionPaths))
}

// ActionPaths indicates the path to where to
// redierect from registration emails
//   - BasePath: fallback, if other paths are not explictily set
//   - ActivationPath: the path to were to redirect when activating
//     a user (the token param will be appended to it)
//   - ResetPasswordPath: the path to were to redirect a user with'
//     a reset password token (the token param will be appended to it)
type ActionPaths struct {
	BasePath          string
	ActivationPath    string
	ResetPasswordPath string
}

// EmailUserAuthRenderData contains the data required
// to render templates tha point to the actual running
// server.
type EmailUserAuthRenderData struct {
	LoginURL                string
	RegisterURL             string
	RequestPasswordResetURL string

	FormErrors []string
}

// NewEmailUserAuthRenderData return the configured urls to redirect a user
// to the login, register or request password web pages
func NewEmailUserAuthRenderData(basePath string) *EmailUserAuthRenderData {
	return &EmailUserAuthRenderData{
		LoginURL:                filepath.Join(basePath, PathLogin),
		RegisterURL:             filepath.Join(basePath, PathRegister),
		RequestPasswordResetURL: filepath.Join(basePath, PathRequestPasswordReset),
	}
}

// HandlerWithEmailRegistration is a handler that takes an extra
// ActionPaths param to be able to complete its operations (like
// sending an email with a password reset url).
type HandlerWithEmailRegistration func(*gin.Context, *ActionPaths)

// emailRegistrationMiddleware wrap a HandlerWithEmailRegistration
// function that gets an actionPaths struct, to convert it to
// a normal gin gonic handler function
func emailRegistrationMiddleware(
	fn HandlerWithEmailRegistration,
	actionPaths ActionPaths) gin.HandlerFunc {
	if len(actionPaths.ActivationPath) == 0 {
		actionPaths.ActivationPath = actionPaths.BasePath + "/" + PathActivate
	}
	if len(actionPaths.ResetPasswordPath) == 0 {
		actionPaths.ResetPasswordPath = actionPaths.BasePath + "/" + PathResetPassword
	}
	return func(c *gin.Context) {
		fn(c, &actionPaths)
	}
}

func buildController(c *gin.Context, actionPaths *ActionPaths) *users.EmailRegistration {
	ed := ginfw.ExtServices(c)
	tokenSalt := fmt.Sprintf("%s%d", c.Request.Host, time.Now().Unix())
	repo := users.NewRepoSQLX(ed.Ins, ed.SQL, tokenSalt)
	// the default is https, and we do not trust the Request Host field
	// because we could be running behind a proxy (like nginx)
	scheme := c.Request.Header.Get("X-Forwarded-Proto")
	if scheme == "" {
		if strings.Contains(c.Request.Host, "localhost") {
			scheme = "http" // for localhost, drop tls
		} else {
			scheme = "https"
		}
	}

	hostInfo := users.HostInfo{
		Scheme:            scheme,
		Host:              c.Request.Host,
		ActivationPath:    actionPaths.ActivationPath,
		ResetPasswordPath: actionPaths.ResetPasswordPath,
	}
	return users.NewEmailRegistration(ed.Ins, ed.Composer, ed.MailSender, repo, hostInfo)
}

// Register implements the user registration api request
func Register(c *gin.Context, actionPaths *ActionPaths) {
	ed := ginfw.ExtServices(c)

	rp := RegisterPayload{}
	err := c.ShouldBindWith(&rp, binding.Form)
	if err != nil {
		ed.Ins.L.Err(err, "failed to bind", map[string]interface{}{
			"error": err.Error(),
		})
		c.HTML(http.StatusOK, TemplActivationSent,
			gin.H{
				"email": "Error",
				"error": err.Error(),
			})
		return
	}

	regUC := buildController(c, actionPaths)
	err = regUC.Register(rp.Email, rp.Password)
	if err != nil {
		if errors.Is(err, users.ErrUserExists) {
			ed.Ins.L.Warn("trying to register user", map[string]interface{}{
				"email": rp.Email,
				"error": err.Error(),
			})
			// we do not send error to avoid leaking existing users email, but
			// we log the attempt
			err = nil
		} else {
			ed.Ins.L.Err(err, "trying to register user", map[string]interface{}{
				"email": rp.Email,
			})
		}
	}
	c.HTML(http.StatusOK, TemplActivationSent,
		gin.H{
			"email": rp.Email,
			"error": err,
		})
}

// RegisterForm returns the template for the registration form
func RegisterForm(c *gin.Context, actionPaths *ActionPaths) {
	c.HTML(http.StatusOK, TemplRegisterForm,
		gin.H{
			"csrf_token":      session.GetCSRFTokenInput(c),
			"email_user_auth": NewEmailUserAuthRenderData(actionPaths.BasePath),
		})
}

// Activate a registered user with the activation token
func Activate(c *gin.Context, actionPaths *ActionPaths) {
	token := c.Query("token")

	if token == "" {
		c.HTML(http.StatusBadRequest, TemplActivateBadToken,
			gin.H{
				"error": "missing token",
			})
		return
	}
	regUC := buildController(c, actionPaths)
	err := regUC.Activate(token)
	if err != nil {
		c.HTML(http.StatusBadRequest, TemplActivateBadToken,
			gin.H{
				"error": err.Error(),
			})
		return
	}
	c.HTML(http.StatusOK, TemplActivateSuccess, gin.H{})
}

// RequestResetPasswordForm renders a form to reset a password with
// a given token
func RequestResetPasswordForm(c *gin.Context, actionPaths *ActionPaths) {
	c.HTML(http.StatusOK, TemplRequestResetPasswordForm,
		gin.H{
			"csrf_token":      session.GetCSRFTokenInput(c),
			"email_user_auth": NewEmailUserAuthRenderData(actionPaths.BasePath),
		})
}

// RequestResetPassword send an email to reset a password
func RequestResetPassword(c *gin.Context, actionPaths *ActionPaths) {
	p := RequestResetPasswordPayload{}
	err := c.ShouldBindWith(&p, binding.Form)
	if err != nil {
		c.HTML(http.StatusOK, TemplResetPasswordTokenSent,
			gin.H{
				"error":           "missing email field",
				"csrf_token":      session.GetCSRFTokenInput(c),
				"email_user_auth": NewEmailUserAuthRenderData(actionPaths.BasePath),
			})
	}
	regUC := buildController(c, actionPaths)
	if err := regUC.RequestResetPassword(p.Email); err != nil {
		// TODO: Instead of TemplResetPasswordTokenSent, create an error
		// template to show what happened.
		c.HTML(http.StatusInternalServerError, TemplResetPasswordTokenSent, gin.H{})
	}
	c.HTML(http.StatusOK, TemplResetPasswordTokenSent, gin.H{})
}

// ResetPasswordForm renders a form to reset a password with
// a given token
func ResetPasswordForm(c *gin.Context, actionPaths *ActionPaths) {
	token := c.Query("token")
	if token == "" {
		c.Redirect(http.StatusFound, actionPaths.BasePath+"/"+PathRequestPasswordReset)
		c.Abort()
		return
	}
	c.HTML(http.StatusOK, TemplResetPasswordForm,
		gin.H{
			"token":           token,
			"csrf_token":      session.GetCSRFTokenInput(c),
			"email_user_auth": NewEmailUserAuthRenderData(actionPaths.BasePath),
		})
}

// ResetPasswordWithToken sets a new password for a user with the password
// reset token sent
func ResetPasswordWithToken(c *gin.Context, actionPaths *ActionPaths) {
	p := ResetPasswordWithTokenPayload{}
	err := c.ShouldBindWith(&p, binding.Form)
	if err != nil {
		deps := ginfw.ExtServices(c)
		deps.Ins.L.
			Warn("cannot bind the payload (redirecting to password reset form)", map[string]interface{}{
				"error": err.Error(),
			})
		c.Redirect(http.StatusFound, actionPaths.BasePath+"/"+PathRequestPasswordReset)
		c.Abort()
		return
	}
	regUC := buildController(c, actionPaths)
	if err := regUC.ResetPasswordWithToken(p.Token, p.Password); err != nil {
		// TODO: Instead of TemplResetPasswordTokenSent, create an error
		// template to show what happened.
		deps := ginfw.ExtServices(c)
		deps.Ins.L.Warn("cannot ResetPasswordWithToken:", map[string]interface{}{
			"token": p.Token,
			"error": err.Error(),
		})
		c.HTML(http.StatusInternalServerError, TemplResetPasswordTokenSent, gin.H{})
	}
	c.HTML(http.StatusOK, TemplResetPasswordSuccess, gin.H{})
}

// LoginForm returns the template to render the login page
func LoginForm(c *gin.Context, actionPaths *ActionPaths) {
	nextPage := c.Query("next_page")
	c.HTML(http.StatusOK, TemplLogin,
		gin.H{
			"csrf_token":      session.GetCSRFTokenInput(c),
			"email_user_auth": NewEmailUserAuthRenderData(actionPaths.BasePath),
			"next_page":       nextPage,
		})
}

// Login executes the login got a given user
func Login(c *gin.Context, actionPaths *ActionPaths) {
	lp := LoginPayload{}
	err := c.ShouldBindWith(&lp, binding.Form)
	ed := ginfw.ExtServices(c)
	if err != nil {
		ed.Ins.L.Err(err, "missing fields", nil)
	}
	regUC := buildController(c, actionPaths)
	userID, err := regUC.Login(lp.Email, lp.Password)
	if err != nil {
		ed.Ins.L.Err(err, "cannot login", nil)
		emailUserAuth := NewEmailUserAuthRenderData(actionPaths.BasePath)
		emailUserAuth.FormErrors = []string{
			"Incorrect email or password",
		}
		c.HTML(http.StatusOK, "wusers_login.html",
			gin.H{
				"csrf_token":      session.GetCSRFTokenInput(c),
				"email_user_auth": emailUserAuth,
			})
	}

	u, _ := regUC.GetUser(userID)
	htmlFields := gin.H{
		"email":           "Not found",
		"created":         "?",
		"email_user_auth": NewEmailUserAuthRenderData(actionPaths.BasePath),
	}
	if u != nil {
		htmlFields["email"] = u.Email
		htmlFields["created"] = u.Created
		session.SetUserID(c, u.ID.ToUUID())
	}
	if len(lp.NextPage) > 0 {
		htmlFields["redirect"] = lp.NextPage
	}
	c.HTML(http.StatusOK, TemplLoginSuccess, htmlFields)
}

// Logout logs out a user
func Logout(c *gin.Context, actionPaths *ActionPaths) {
	htmlFields := gin.H{
		"email":           "Not found",
		"created":         "?",
		"email_user_auth": NewEmailUserAuthRenderData(actionPaths.BasePath),
	}
	session.ClearUserID(c)
	c.HTML(http.StatusOK, TemplLogoutSuccess, htmlFields)
}
