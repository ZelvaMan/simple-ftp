package ftp

import "fmt"

type ServerError struct {
	msg              string
	respondMessage   string
	statusCode       int
	terminateSession bool
}

func NewError(msg string, respondMessage string, statusCode int, terminateSession bool) *ServerError {
	return &ServerError{
		msg:              msg,
		statusCode:       statusCode,
		terminateSession: terminateSession,
	}
}

func (e *ServerError) Error() string {
	return fmt.Sprintf("%d %s", e.statusCode, e.respondMessage)
}

func (e *ServerError) ShouldTerminate() bool {
	return e.terminateSession
}
