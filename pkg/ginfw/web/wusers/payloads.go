package wusers

// RegisterPayload contains the data to register a new user.
type RegisterPayload struct {
	Email            string `form:"email" binding:"required"`
	Password         string `form:"password1" binding:"required"`
	PasswordRepeated string `form:"password2" binding:"required"`
}

// LoginPayload contains the data required to log in a user.
type LoginPayload struct {
	Email    string `form:"username" binding:"required"`
	Password string `form:"password" binding:"required"`
	NextPage string `form:"nextpage"`
}

// RequestResetPasswordPayload contains the data to request
// a password reset for a user.
type RequestResetPasswordPayload struct {
	Email string `form:"email" binding:"required"`
}

// ResetPasswordWithTokenPayload contains the data to create
// a new password for a user using a reset password token.
type ResetPasswordWithTokenPayload struct {
	Token    string `form:"token" binding:"required"`
	Password string `form:"password" binding:"required"`
}
