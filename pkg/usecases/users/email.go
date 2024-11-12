package users

import (
	"fmt"

	"github.com/dhontecillas/hfw/pkg/ids"
	"github.com/dhontecillas/hfw/pkg/mailer"
	"github.com/dhontecillas/hfw/pkg/notifications"
	"github.com/dhontecillas/hfw/pkg/obs"
)

// Notification template names
const (
	NotifRequestRegistration  string = "users_requestregistration"
	NotifRequestPasswordReset string = "users_requestpasswordreset"

	carrierEmail = "email"
)

// RegistrationRepo defines the data access interface to
// implement an Email user registration (and reset password)
// workflow.
type RegistrationRepo interface {

	// GetUserByEmail returns the User for a given email,
	// or nil if the email is not found in the data repo.
	GetUserByEmail(email string) *User

	// GetUserByID returns a User for a given ID, or nil
	// if there is no user for that ID.
	GetUserByID(userID ids.ID) *User

	// CreateInactiveUser return a token for the user to be used
	// to confirm the account.
	CreateInactiveUser(email string, password string) (string, error)

	// ActivateUser confirms an user email with its activation token.
	ActivateUser(token string) (*User, error)

	// CreatePasswordResetRequest returns a token for password reset.
	CreatePasswordResetRequest(email string) (*User, string, error)

	// ResetPassword changes the pasword for the user associated with
	// the given reset password token
	ResetPassword(token string, password string) (*User, error)

	// CheckPassword return the user ID for a user from its email
	// and password.
	CheckPassword(email string, password string) (ids.ID, error)

	// DeleteUser deletes a created user
	DeleteUser(email string) error

	// ListUsers lists users with pagination
	ListUsers(from ids.ID, limit int, backwards bool) ([]User, error)
}

// HostInfo contains the required info to construct
// a URL to where a user can be redirected to use an
// activation token, or use a reset password token.
// This information might not be available if the request
// is proxied, so it can be configured at startup time.
type HostInfo struct {
	Scheme            string // http or https
	Host              string // includes port
	ActivationPath    string // activation path where the handler has been installed
	ResetPasswordPath string // reset pass path where the handler has been installed
}

// EmailRegistration is the controller to handle
// an email registration flow
type EmailRegistration struct {
	ins        *obs.Insighter
	composer   notifications.Composer
	mailSender mailer.Mailer
	regRepo    RegistrationRepo
	hostInfo   HostInfo
}

// NewEmailRegistration creates a new EmailRegistration
// controller
func NewEmailRegistration(
	ins *obs.Insighter,
	composer notifications.Composer,
	mailSender mailer.Mailer,
	regRepo RegistrationRepo,
	hostInfo HostInfo) *EmailRegistration {
	return &EmailRegistration{
		ins:        ins,
		composer:   composer,
		mailSender: mailSender,
		regRepo:    regRepo,
		hostInfo:   hostInfo,
	}
}

func (r *EmailRegistration) sendMail(email string, notification string,
	data map[string]interface{}) error {

	content, err := r.composer.Render(notification, data, carrierEmail)
	if err != nil {
		// TODO: wrap the error here
		return err
	}

	subject, ok := content.Texts["subject"]
	if !ok {
		return fmt.Errorf("missing 'subject' content")
	}
	textBody, ok := content.Texts["content"]
	if !ok {
		return fmt.Errorf("missing 'text body' content")
	}
	htmlBody, ok := content.HTMLs["content"]
	if !ok {
		return fmt.Errorf("missing 'html body' content")
	}

	fromName, fromAddress := r.mailSender.Sender()
	emailMessage := mailer.Email{
		To: mailer.User{
			Name:    email,
			Address: email,
		},
		From: mailer.User{
			Name:    fromName,
			Address: fromAddress,
		},
		Subject: subject,
		HTML:    htmlBody,
		Text:    textBody,
	}

	if sendEmailErr := r.mailSender.Send(emailMessage); err != nil {
		r.ins.L.Err(sendEmailErr, "cannot send registration message")
		return fmt.Errorf("%w %s", ErrNotificationFailed, sendEmailErr)
	}
	return nil
}

// Register creates an inactive user, and send a notification (email),
// with the activation link.
func (r *EmailRegistration) Register(email string, password string) error {
	token, e := r.regRepo.CreateInactiveUser(email, password)
	if e != nil {
		// TODO: if the user is already in the database, we might have
		// the issue that is in activation pending, because it failed
		// to send the notification, so, in that case, we could retry
		// to send an activation link to the email.
		r.ins.L.Err(e, "cannot create innactive user")
		return fmt.Errorf("cannot create innactive user: %w", e)
	}

	return r.sendMail(email, NotifRequestRegistration, map[string]interface{}{
		"to_address":       email,
		"activation_token": token,
		"scheme":           r.hostInfo.Scheme,
		"host":             r.hostInfo.Host,
		"path":             r.hostInfo.ActivationPath,
	})
}

// Activate completes the activation of a registered user.
func (r *EmailRegistration) Activate(token string) error {
	_, err := r.regRepo.ActivateUser(token)
	if err != nil {
		return err
	}
	return nil
}

// RequestResetPassword creates a temporal reset token and sends
// it to the user email, so it can reset the password.
func (r *EmailRegistration) RequestResetPassword(email string) error {
	u, token, err := r.regRepo.CreatePasswordResetRequest(email)
	if err != nil {
		return err
	}

	return r.sendMail(u.Email, NotifRequestPasswordReset,
		map[string]interface{}{
			"to_address": u.Email,
			"token":      token,
			"scheme":     r.hostInfo.Scheme,
			"host":       r.hostInfo.Host,
			"path":       r.hostInfo.ResetPasswordPath,
		})
}

// ResetPasswordWithToken sets a new password for a given user using a
// reset password token
func (r *EmailRegistration) ResetPasswordWithToken(token string, newPassword string) error {
	_, err := r.regRepo.ResetPassword(token, newPassword)
	if err != nil {
		return err
	}
	return nil
}

// Login check if a user email and password are correct.
func (r *EmailRegistration) Login(email string, password string) (ids.ID, error) {
	return r.regRepo.CheckPassword(email, password)
}

// GetUser returns a user given its ID
func (r *EmailRegistration) GetUser(userID ids.ID) (*User, error) {
	u := r.regRepo.GetUserByID(userID)
	if u == nil {
		return nil, ErrNotFound
	}
	return u, nil
}
