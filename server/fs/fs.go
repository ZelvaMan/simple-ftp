package fs

import (
	"io"
)

type Filesystem interface {
	List(directory string) ([]File, error)
	Retrieve(path string) (io.Reader, error)
}
