package updateclihttp

import "fmt"

type ErrHttpError struct {
	resStatusCode int
}

func (e *ErrHttpError) Error() string {
	return fmt.Sprintf("[http] status code %d received while 1XX, 2XX or 3XX is expected.", e.resStatusCode)
}
