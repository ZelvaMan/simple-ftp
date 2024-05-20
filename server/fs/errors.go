package fs

import "fmt"

type NotFoundError struct {
	path string
}

func (e NotFoundError) Error() string {
	return fmt.Sprintf("file %s not found", e.path)
}

func NewNotFoundError(path string) NotFoundError {
	return NotFoundError{path: path}
}
