package fs

import (
	"fmt"
	"time"
)

type File struct {
	Name         string
	Size         int64
	LastModified time.Time
	IsDir        bool
	Permissions  string
}

func (fileInfo File) String() string {
	modifiedFormatted := fileInfo.LastModified.Format("Jan 02 03:04")
	return fmt.Sprintf("%s  1 peter         %d %s %s \n", fileInfo.Permissions, fileInfo.Size, modifiedFormatted, fileInfo.Name)
}
