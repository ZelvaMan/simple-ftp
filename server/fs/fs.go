package fs

import (
	"io"
)

type Filesystem interface {
	List(directory string) (FileList, error)
	Retrieve(path string) (io.Reader, error)
	Store(path string, data io.Reader) error
}
