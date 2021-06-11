package errors

import (
	"errors"

	"github.com/davidebianchi/go-jsonclient"
)

var (
	// ErrGeneric is a generic error
	ErrGeneric = errors.New("Something went wrong")
	// ErrHTTP is an http error
	ErrHTTP = jsonclient.ErrHTTP
	// ErrCreateClient is the error creating MiaSdkClient
	ErrCreateClient = errors.New("Error creating sdk client")
	// ErrProjectNotFound is the error returned when specified
	// project cannot be found.
	ErrProjectNotFound = errors.New("Project not found")
)
