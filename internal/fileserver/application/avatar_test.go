package application

import (
	"context"
	"errors"
	"io"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
)

type fakeStorage struct {
	gotExt  string
	gotData []byte
	name    string
	err     error
}

func (f *fakeStorage) Save(_ context.Context, data io.Reader, extension string) (string, error) {
	if f.err != nil {
		return "", f.err
	}
	f.gotExt = extension
	b, err := io.ReadAll(data)
	if err != nil {
		return "", err
	}
	f.gotData = b
	return f.name, nil
}

func TestAvatarService_Upload(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name        string
		data        []byte
		extension   string
		storageName string
		storageErr  error
		wantErr     error
		wantName    string
		wantExt     string
	}{
		{
			name:        "success png",
			data:        []byte("hello"),
			extension:   ".png",
			storageName: "abc.png",
			wantName:    "abc.png",
			wantExt:     ".png",
		},
		{
			name:        "extension without dot is normalized",
			data:        []byte("hello"),
			extension:   "jpg",
			storageName: "abc.jpg",
			wantName:    "abc.jpg",
			wantExt:     ".jpg",
		},
		{
			name:        "empty extension defaults to .bin",
			data:        []byte("hello"),
			extension:   "",
			storageName: "abc.bin",
			wantName:    "abc.bin",
			wantExt:     ".bin",
		},
		{
			name:      "empty data is rejected",
			data:      nil,
			extension: ".png",
			wantErr:   ErrEmptyData,
		},
		{
			name:      "too large is rejected",
			data:      make([]byte, MaxUploadBytes+1),
			extension: ".png",
			wantErr:   ErrTooLarge,
		},
		{
			name:       "storage error is propagated",
			data:       []byte("x"),
			extension:  ".png",
			storageErr: errors.New("disk full"),
			wantErr:    errors.New("disk full"),
		},
	}

	for _, c := range cases {
		c := c
		t.Run(c.name, func(t *testing.T) {
			t.Parallel()

			fs := &fakeStorage{name: c.storageName, err: c.storageErr}
			svc := NewAvatarService(fs)

			name, err := svc.Upload(context.Background(), c.data, c.extension)

			if c.wantErr != nil {
				require.Error(t, err)
				if errors.Is(c.wantErr, ErrEmptyData) || errors.Is(c.wantErr, ErrTooLarge) {
					require.ErrorIs(t, err, c.wantErr)
				}
				return
			}
			require.NoError(t, err)
			require.Equal(t, c.wantName, name)
			require.Equal(t, c.wantExt, fs.gotExt)
			require.Equal(t, strings.Repeat("hello", 1), string(fs.gotData))
		})
	}
}

func TestAvatarService_NormalizeExtension(t *testing.T) {
	t.Parallel()

	require.Equal(t, ".bin", normalizeExtension(""))
	require.Equal(t, ".bin", normalizeExtension("   "))
	require.Equal(t, ".png", normalizeExtension(".png"))
	require.Equal(t, ".png", normalizeExtension("png"))
}
