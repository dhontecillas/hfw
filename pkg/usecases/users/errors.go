package users

import (
	"github.com/dhontecillas/hfw/pkg/consterr"
)

// Error definitions for the user registration, login
// and reset password flows.
const (
	ErrUserExists = consterr.ConstErr("ErrUserExists")
	ErrNotFound   = consterr.ConstErr("ErrNotFound")
	ErrConsumed   = consterr.ConstErr("ErrConsumed")
	ErrExpired    = consterr.ConstErr("ErrExpired")

	ErrNotificationFailed = consterr.ConstErr("ErrNotificationFailed")
)
