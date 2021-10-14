package ws

import (
	"fmt"
	"net/http"
)

type ErrAuth struct {
	HTTPStatus int
	Message    string
}

func (e ErrAuth) Error() string {
	return fmt.Sprintf("%s (code %d)", e.Message, e.HTTPStatus)
}

type Authenticator interface {
	// Authenticate takes a request, authenticates it and returns a manager that
	// will handle the request.
	// If err is of type ErrAuth, then the HTTP response is returned with
	// ErrAuth.HTTPStatus code and following JSON body:
	// { "message" : "ErrAuth.Message" }
	// If err is of any other type, the error is logged and a 500 code is
	// returned to the client.
	Authenticate(r *http.Request) (*Manager, error)
}

type DefaultAuthenticator struct {
	Manager *Manager
}

func (d *DefaultAuthenticator) Authenticate(_ *http.Request) (*Manager, error) {
	return d.Manager, nil
}
