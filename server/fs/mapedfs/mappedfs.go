package mapedfs

import (
	"errors"
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"
	"server/fs"
)

// MappedFS implements Filesystem
type MappedFS struct {
	osFSRoot string // = / in mapped fs
}

func CreateFS(osRoot string) (*MappedFS, error) {
	// TODO check if root exists and is folder
	filesystem := MappedFS{osFSRoot: osRoot}

	return &filesystem, nil
}

func (mfs *MappedFS) List(directory string) (fs.FileList, error) {
	realPath := mfs.resolveMappedToReal(directory)

	entries, err := os.ReadDir(realPath)
	if err != nil {
		return nil, fmt.Errorf("error reading reading directory in folder %s(%s): %s", directory, realPath, err)
	}

	mappedEntries := make([]fs.File, len(entries))

	for idx, value := range entries {
		info, err := value.Info()
		if err != nil {
			return nil, fmt.Errorf("error reading info of file %s : %s", value.Name(), err)
		}

		var enhancedPermission = ""
		if info.IsDir() {
			enhancedPermission = "d" + info.Mode().Perm().String()
		} else {
			enhancedPermission = "-" + info.Mode().Perm().String()
		}

		mappedEntries[idx] = fs.File{
			Name:         value.Name(),
			IsDir:        value.IsDir(),
			Size:         info.Size(),
			LastModified: info.ModTime(),
			Permissions:  enhancedPermission,
		}
	}

	return mappedEntries, nil

}

func (mfs *MappedFS) Retrieve(path string) (io.Reader, error) {
	realPath := mfs.resolveMappedToReal(path)

	file, err := os.Open(realPath)
	if errors.Is(err, os.ErrNotExist) {
		return nil, fs.NewNotFoundError(path)
	}
	if err != nil {
		return nil, fmt.Errorf("retrieve file %s:%s", path, err)
	}

	log.Printf("File reader received for file %s(%s)", path, realPath)
	return file, nil
}

func (mfs *MappedFS) Store(path string, data io.Reader) error {
	realPath := mfs.resolveMappedToReal(path)
	file, err := os.OpenFile(realPath, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0777)

	defer func() {
		err := file.Close()
		if err != nil {
			log.Printf("Eror closing file: %s", err)

		}
	}()

	if err != nil {
		return fmt.Errorf("opening file for writing: %s", err)
	}

	log.Printf("file opened, starting to copy data")
	_, err = io.Copy(file, data)
	if err != nil {
		return fmt.Errorf("copying from data to file: %s", err)
	}

	log.Printf("data copied")

	return nil

}

func (mfs *MappedFS) Exists(path string) (bool, error) {
	realPath := mfs.resolveMappedToReal(path)
	_, err := os.Stat(realPath)
	if err == nil {
		return true, nil
	}
	if os.IsNotExist(err) {
		return false, nil
	}

	return false, fmt.Errorf("mapped fs error: %s", err)

}

func (mfs *MappedFS) Rename(from string, to string) error {
	realFrom := mfs.resolveMappedToReal(from)
	realTo := mfs.resolveMappedToReal(to)

	err := os.Rename(realFrom, realTo)
	if err != nil {
		return fmt.Errorf("mapped fs error: %s", err)
	}

	log.Printf("MappedOS: file %s renamed to %s", from, to)
	return nil
}

func (mfs *MappedFS) Delete(deletePath string) error {
	realPath := mfs.resolveMappedToReal(deletePath)

	err := os.Remove(realPath)
	if err != nil {
		return fmt.Errorf("mapped fs error: %s", err)
	}

	log.Printf("MappedFS: file %s deleted", deletePath)

	return nil
}

func (mfs *MappedFS) CreateDirectory(directory string) error {
	realPath := mfs.resolveMappedToReal(directory)

	err := os.Mkdir(realPath, 0777)
	if err != nil {
		return fmt.Errorf("mapped fs error: %s", err)
	}

	log.Printf("MappedFS: directory %s created", directory)

	return nil
}

func (mfs *MappedFS) resolveMappedToReal(relativePath string) string {
	// ensures that file that is not inside osFSRoot is permitted
	clearedPath := filepath.Clean(relativePath)
	realPath := filepath.Join(mfs.osFSRoot, clearedPath)
	log.Printf("rel filepath %s resolved to %s", relativePath, realPath)

	return realPath
}
