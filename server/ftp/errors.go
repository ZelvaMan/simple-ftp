package ftp

import "fmt"

func respondError(command string, err error) error {
	return fmt.Errorf("sending response for command %s: %s", command, err)
}
