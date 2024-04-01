package ftp

import "fmt"

func formatResponse(responseCode int, message string) string {
	return fmt.Sprintf("%d %s", responseCode, message)
}
