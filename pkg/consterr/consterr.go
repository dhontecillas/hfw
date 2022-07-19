package consterr

// ConstErr is a basic error type from a string.
type ConstErr string

// Error implements the error interface for ConstErr
func (e ConstErr) Error() string {
	return string(e)
}
