package fs

import (
	"fmt"
	"strings"
	"time"
)

type File struct {
	Name         string
	Size         int64
	LastModified time.Time
	IsDir        bool
	Permissions  string
}
type FileList []File

func (fileInfo File) String() string {
	modifiedFormatted := fileInfo.LastModified.Format("Jan 02 03:04")
	return fmt.Sprintf("%s  1 peter         %d %s %s \n", fileInfo.Permissions, fileInfo.Size, modifiedFormatted, fileInfo.Name)
}

func (files FileList) String() string {
	var builder strings.Builder

	for _, file := range files {
		builder.WriteString(fmt.Sprintf("%s\r\n", file.String()))
	}

	return builder.String()
}
