package application

import (
	"bytes"
	"context"
	"errors"
	"io"
	"strings"
)

const MaxUploadBytes = 6 << 20

var ErrTooLarge = errors.New("application: file is too large")
var ErrEmptyData = errors.New("application: empty data")

//go:generate go run go.uber.org/mock/mockgen@latest -source=avatar.go -destination=mocks/mock_avatar.go -package=mocks
type Storage interface {
	Save(ctx context.Context, data io.Reader, extension string) (string, error)
}

type AvatarService struct {
	storage Storage
}

func NewAvatarService(storage Storage) *AvatarService {
	return &AvatarService{storage: storage}
}

func normalizeExtension(extension string) string {
	ext := strings.TrimSpace(extension)
	if ext == "" {
		return ".bin"
	}
	if !strings.HasPrefix(ext, ".") {
		ext = "." + ext
	}
	return ext
}

func (s *AvatarService) Upload(ctx context.Context, data []byte, extension string) (string, error) {
	if len(data) == 0 {
		return "", ErrEmptyData
	}
	if len(data) > MaxUploadBytes {
		return "", ErrTooLarge
	}
	ext := normalizeExtension(extension)
	return s.storage.Save(ctx, bytes.NewReader(data), ext)
}
