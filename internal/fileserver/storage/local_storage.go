package storage

import (
	"context"
	"errors"
	"io"
	"os"
	"path/filepath"

	"github.com/google/uuid"
)

var ErrEmptyName = errors.New("storage: empty filename")

type LocalStorage struct {
	dir string
}

func NewLocalStorage(dir string) *LocalStorage {
	return &LocalStorage{dir: dir}
}

func (s *LocalStorage) Save(ctx context.Context, data io.Reader, extension string) (string, error) {
	if err := ctx.Err(); err != nil {
		return "", err
	}
	if err := os.MkdirAll(s.dir, 0o755); err != nil {
		return "", err
	}

	name := filepath.Base(uuid.New().String() + extension)
	if name == "" || name == "." || name == "/" {
		return "", ErrEmptyName
	}
	dst, err := os.OpenFile(filepath.Join(s.dir, name), os.O_CREATE|os.O_WRONLY|os.O_EXCL, 0o644)
	if err != nil {
		return "", err
	}
	defer dst.Close()

	if _, err := io.Copy(dst, data); err != nil {
		_ = os.Remove(filepath.Join(s.dir, name))
		return "", err
	}
	return name, nil
}
