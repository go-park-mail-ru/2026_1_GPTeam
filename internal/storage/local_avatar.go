package storage

import (
	"context"
	"io"
	"os"
	"path/filepath"

	"github.com/google/uuid"
)

type LocalAvatar struct {
	dir string
}

func NewLocalAvatar(dir string) *LocalAvatar {
	return &LocalAvatar{dir: dir}
}

func (s *LocalAvatar) Upload(ctx context.Context, reader io.Reader, extension string) (string, error) {
	_ = ctx
	if err := os.MkdirAll(s.dir, 0755); err != nil {
		return "", err
	}
	name := uuid.New().String() + extension
	dst, err := os.Create(filepath.Join(s.dir, filepath.Base(name)))
	if err != nil {
		return "", err
	}
	defer dst.Close()
	if _, err := io.Copy(dst, reader); err != nil {
		return "", err
	}
	return filepath.Base(name), nil
}
