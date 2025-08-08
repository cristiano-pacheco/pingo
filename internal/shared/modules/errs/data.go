package errs

import "fmt"

type Error struct {
	Status        int   `json:"-"`
	OriginalError error `json:"-"`
	Err           er    `json:"error"`
}

type er struct {
	Code    string   `json:"code"`
	Message string   `json:"message"`
	Details []detail `json:"details,omitempty"`
}

type detail struct {
	Field   string `json:"field,omitempty"`
	Message string `json:"message,omitempty"`
}

func (e *Error) Error() string {
	msg := fmt.Sprintf("[%s] %s", e.Err.Code, e.Err.Message)
	return msg
}
