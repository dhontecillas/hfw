package users

import (
	"fmt"

	"github.com/dhontecillas/hfw/pkg/ids"
)

var _ RegistrationRepo = (*NopRegistrationRepo)(nil)

// NopRegistrationRepo is an empty not working
// implementation of a RegistrationRepo
type NopRegistrationRepo struct {
}

// GetUserByEmail returns the User for a given email,
// or nil if the email is not found in the data repo.
func (r *NopRegistrationRepo) GetUserByEmail(email string) *User {
	return nil
}

// GetUserByID returns a User for a given ID, or nil
// if there is no user for that ID.
func (r *NopRegistrationRepo) GetUserByID(userID ids.ID) *User {
	return nil
}

// CreateInactiveUser return a token for the user to be used
// to confirm the account.
func (r *NopRegistrationRepo) CreateInactiveUser(
	email string, password string) (string, error) {
	return "", fmt.Errorf("not implemented")
}

// ActivateUser confirms an user email with its activation token.
func (r *NopRegistrationRepo) ActivateUser(token string) (*User, error) {
	return nil, fmt.Errorf("not implemented")
}

// CreatePasswordResetRequest returns a token for password reset.
func (r *NopRegistrationRepo) CreatePasswordResetRequest(email string) (*User, string, error) {
	return nil, "", fmt.Errorf("not implemented")
}

// ResetPassword changes the pasword for the user associated with
// the given reset password token
func (r *NopRegistrationRepo) ResetPassword(token string, password string) (*User, error) {
	return nil, fmt.Errorf("not implemented")
}

// CheckPassword return the user ID for a user from its email
// and password.
func (r *NopRegistrationRepo) CheckPassword(email string, password string) (ids.ID, error) {
	var id ids.ID
	return id, fmt.Errorf("not implemented")
}

// DeleteUser deletes an existing user
func (r *NopRegistrationRepo) DeleteUser(email string) error {
	return fmt.Errorf("not implemented")
}
