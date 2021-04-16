package vote

import "fmt"

type typeError struct {
	t   string
	err error
}

func (c typeError) Type() string {
	return c.t
}

func (c typeError) Error() string {
	return fmt.Sprintf(`{"error":"%s","msg":"%s"}`, c.t, c.err)
}

func (c typeError) Unwrap() error {
	return c.err
}
