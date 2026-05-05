package grpcserver

import (
	"context"
	"errors"
	"testing"

	"github.com/go-park-mail-ru/2026_1_GPTeam/internal/fileserver/application"
	fsv1 "github.com/go-park-mail-ru/2026_1_GPTeam/pkg/gen/fileserver/v1"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type stubUseCase struct {
	gotData []byte
	gotExt  string
	name    string
	err     error
}

func (s *stubUseCase) Upload(_ context.Context, data []byte, extension string) (string, error) {
	s.gotData = data
	s.gotExt = extension
	return s.name, s.err
}

func TestServer_Upload(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name      string
		req       *fsv1.UploadRequest
		usecase   *stubUseCase
		wantCode  codes.Code
		wantName  string
		expectErr bool
	}{
		{
			name:     "success",
			req:      &fsv1.UploadRequest{Data: []byte("img"), Extension: ".png"},
			usecase:  &stubUseCase{name: "abc.png"},
			wantName: "abc.png",
		},
		{
			name:      "empty data is invalid argument",
			req:       &fsv1.UploadRequest{Data: nil, Extension: ".png"},
			usecase:   &stubUseCase{},
			wantCode:  codes.InvalidArgument,
			expectErr: true,
		},
		{
			name:      "too large from app maps to invalid argument",
			req:       &fsv1.UploadRequest{Data: []byte("x"), Extension: ".png"},
			usecase:   &stubUseCase{err: application.ErrTooLarge},
			wantCode:  codes.InvalidArgument,
			expectErr: true,
		},
		{
			name:      "empty data from app maps to invalid argument",
			req:       &fsv1.UploadRequest{Data: []byte("x"), Extension: ".png"},
			usecase:   &stubUseCase{err: application.ErrEmptyData},
			wantCode:  codes.InvalidArgument,
			expectErr: true,
		},
		{
			name:      "unknown error becomes internal",
			req:       &fsv1.UploadRequest{Data: []byte("x"), Extension: ".png"},
			usecase:   &stubUseCase{err: errors.New("boom")},
			wantCode:  codes.Internal,
			expectErr: true,
		},
	}

	for _, c := range cases {
		c := c
		t.Run(c.name, func(t *testing.T) {
			t.Parallel()
			srv := NewServer(c.usecase)

			resp, err := srv.Upload(context.Background(), c.req)

			if c.expectErr {
				require.Error(t, err)
				st, ok := status.FromError(err)
				require.True(t, ok)
				require.Equal(t, c.wantCode, st.Code())
				return
			}
			require.NoError(t, err)
			require.Equal(t, c.wantName, resp.GetFilename())
			require.Equal(t, ".png", c.usecase.gotExt)
			require.Equal(t, []byte("img"), c.usecase.gotData)
		})
	}
}
