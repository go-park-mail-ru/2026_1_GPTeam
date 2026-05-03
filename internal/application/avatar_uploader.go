package application

import (
	"context"
	"io"
)

type AvatarUploader interface {
	Upload(ctx context.Context, reader io.Reader, extension string) (filename string, err error)
}
