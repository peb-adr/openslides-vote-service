package vote

import "fmt"

const (
	// ErrUnknown should not happen.
	ErrUnknown TypeError = iota

	// ErrExists happens, when start is called with an poll ID that already
	// exists.
	ErrExists

	// ErrNotExists happens when an operation is performed on an unknown poll.
	ErrNotExists

	// ErrInvalid happens, when the vote data is invalid.
	ErrInvalid
)

// TypeError is an error that can happend in this API.
type TypeError int

// Type returns a name for the error.
func (err TypeError) Type() string {
	switch err {
	case ErrExists:
		return "exist"

	case ErrNotExists:
		return "not-exist"

	case ErrInvalid:
		return "invalid"

	default:
		return "unknown"
	}
}

func (err TypeError) Error() string {
	var msg string
	switch err {
	case ErrExists:
		msg = "Poll does already exist with differet config"

	case ErrNotExists:
		msg = "Poll does not exist"

	case ErrInvalid:
		msg = "The input data is invalid"

	default:
		msg = "Unknown error"

	}
	return fmt.Sprintf(`{"error":"%s","msg":"%s"}`, err.Type(), msg)
}

// MessageError is a TypeError with an individuel error message.
type MessageError struct {
	TypeError
	msg string
}

func (err MessageError) Error() string {
	return fmt.Sprintf(`{"error":"%s","msg":"%s"}`, err.Type(), err.msg)
}

func (err MessageError) Unwrap() error {
	return err.TypeError
}
