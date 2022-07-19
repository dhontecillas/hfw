package wusers

// These are the definitions for paths and template names
// for user registration flows.
const (
	PathRegister             string = "register"
	PathActivate             string = "activate"
	PathLogin                string = "login"
	PathLogout               string = "logout"
	PathRequestPasswordReset string = "requestresetpassword"
	PathResetPassword        string = "resetpassword"

	TemplLogin string = "wusers_login.html"

	TemplRegisterForm     string = "wusers_register.html"
	TemplActivationSent   string = "wusers_registration_activation_sent.html"
	TemplActivateBadToken string = "wusers_registration_activation_bad_token.html"
	TemplActivateSuccess  string = "wusers_registration_activation_success.html"

	TemplLoginForm     string = "wusers_login.html"
	TemplLoginSuccess  string = "wusers_login_completed.html"
	TemplLogoutSuccess string = "wusers_login_completed.html"

	TemplRequestResetPasswordForm string = "wusers_registration_request_password_reset_form.html"
	TemplResetPasswordTokenSent   string = "wusers_registration_reset_password_token_sent.html"
	TemplResetPasswordForm        string = "wusers_registration_reset_password_form.html"
	TemplResetPasswordSuccess     string = "wusers_registration_reset_password_success.html"
)
