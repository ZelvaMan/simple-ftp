package respones

import "fmt"

func formatResponse(responseCode int, message string) string {
	return fmt.Sprintf("%d %s", responseCode, message)
}

func UserLoggedIn() string {
	return formatResponse(230, "User logged in, proceed.")
}

func PasswordNeeded() string {
	return formatResponse(331, "User name okay, need password.")
}

func BadSequence() string {
	return formatResponse(503, "Bad sequence of commands.")
}
func Ready() string {
	return formatResponse(220, "zmftp ready for new user.")
}

func NotLoggedIn() string {
	return formatResponse(530, "Not logged in / incorrect password.")
}

func NotImplemented() string {
	return formatResponse(502, "Command not implemented.")
}
