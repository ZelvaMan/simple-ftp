package ftp

import "fmt"

func formatResponse(responseCode int, message string) string {
	return fmt.Sprintf("%d %s", responseCode, message)
}

func userLoggedInResponse() string {
	return formatResponse(230, "User logged in, proceed.")
}
