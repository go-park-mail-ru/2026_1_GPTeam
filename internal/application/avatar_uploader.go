package application

import (
	"context"
	"io"
)

//go:generate go run go.uber.org/mock/mockgen@latest -source=avatar_uploader.go -destination=mocks/mock_avatar_uploader.go -package=mocks
type AvatarUploader interface {
	Upload(ctx context.Context, reader io.Reader, extension string) (filename string, err error)
}
