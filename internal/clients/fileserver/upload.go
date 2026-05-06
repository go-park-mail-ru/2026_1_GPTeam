package fileserver

import (
	"context"
	"fmt"
	"io"

	fsv1 "github.com/go-park-mail-ru/2026_1_GPTeam/pkg/gen/fileserver/v1"
)

const MaxClientReadBytes = 8 << 20

type GrpcUploader struct {
	client fsv1.FileServiceClient
}

func NewGrpcUploader(client fsv1.FileServiceClient) *GrpcUploader {
	return &GrpcUploader{client: client}
}

func (c *GrpcUploader) Upload(ctx context.Context, reader io.Reader, extension string) (string, error) {
	data, err := io.ReadAll(io.LimitReader(reader, MaxClientReadBytes))
	if err != nil {
		return "", fmt.Errorf("fileserver: read body: %w", err)
	}
	resp, err := c.client.Upload(ctx, &fsv1.UploadRequest{
		Data:      data,
		Extension: extension,
	})
	if err != nil {
		return "", fmt.Errorf("fileserver: upload rpc: %w", err)
	}
	if resp.GetFilename() == "" {
		return "", fmt.Errorf("fileserver: empty filename")
	}
	return resp.GetFilename(), nil
}
