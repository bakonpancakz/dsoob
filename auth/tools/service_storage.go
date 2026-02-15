package tools

import (
	"errors"
	"io"
	"os"
	"path"
)

var (
	ErrStorageFileNotFound    = errors.New("file not found")
	ErrStorageInvalidFilename = errors.New("filename contains invalid characters")
)

// Create a public file in the `{DATA_DIRECTORY}/public/{filepath}{filename}` directory
func StoragePublicCreate(filepath, filename string, perms os.FileMode) (io.WriteCloser, error) {
	fd := path.Join(DATA_DIRECTORY, "public", filepath)
	fp := path.Join(fd, filename)

	if err := os.MkdirAll(fd, perms); err != nil {
		return nil, err
	}

	return os.OpenFile(fp, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, perms)
}

// Delete a public file in the `{DATA_DIRECTORY}/public/` directory
func StoragePublicDelete(filenames ...string) error {
	for _, fn := range filenames {
		fp := path.Join(DATA_DIRECTORY, "public", fn)
		if err := os.Remove(fp); err != nil {
			return err
		}
	}
	return nil
}
