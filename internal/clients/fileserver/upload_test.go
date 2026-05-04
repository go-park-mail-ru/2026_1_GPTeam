package fileserver

import (
	"context"
	"errors"
	"io"
	"net"
	"strings"
	"testing"
	"time"

	fsv1 "github.com/go-park-mail-ru/2026_1_GPTeam/pkg/gen/fileserver/v1"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/status"
)

type fakeServer struct {
	fsv1.UnimplementedFileServiceServer
	gotData      []byte
	gotExtension string
	respName     string
	respErr      error
}

func (f *fakeServer) Upload(_ context.Context, req *fsv1.UploadRequest) (*fsv1.UploadResponse, error) {
	if f.respErr != nil {
		return nil, f.respErr
	}
	f.gotData = req.GetData()
	f.gotExtension = req.GetExtension()
	return &fsv1.UploadResponse{Filename: f.respName}, nil
}

func startTestServer(t *testing.T, srv *fakeServer) (fsv1.FileServiceClient, func()) {
	t.Helper()

	lis, err := net.Listen("tcp", "127.0.0.1:0")
	require.NoError(t, err)

	g := grpc.NewServer()
	fsv1.RegisterFileServiceServer(g, srv)

	go func() { _ = g.Serve(lis) }()

	conn, err := grpc.NewClient(
		lis.Addr().String(),
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	)
	require.NoError(t, err)

	cleanup := func() {
		_ = conn.Close()
		g.Stop()
	}
	return fsv1.NewFileServiceClient(conn), cleanup
}

func TestGrpcUploader_Success(t *testing.T) {
	t.Parallel()

	fake := &fakeServer{respName: "abc.png"}
	client, cleanup := startTestServer(t, fake)
	t.Cleanup(cleanup)

	uploader := NewGrpcUploader(client)
	name, err := uploader.Upload(context.Background(), strings.NewReader("payload"), ".png")

	require.NoError(t, err)
	require.Equal(t, "abc.png", name)
	require.Equal(t, []byte("payload"), fake.gotData)
	require.Equal(t, ".png", fake.gotExtension)
}

func TestGrpcUploader_EmptyFilenameIsError(t *testing.T) {
	t.Parallel()

	fake := &fakeServer{respName: ""}
	client, cleanup := startTestServer(t, fake)
	t.Cleanup(cleanup)

	uploader := NewGrpcUploader(client)
	_, err := uploader.Upload(context.Background(), strings.NewReader("payload"), ".png")
	require.Error(t, err)
	require.Contains(t, err.Error(), "empty filename")
}

func TestGrpcUploader_RpcErrorIsWrapped(t *testing.T) {
	t.Parallel()

	fake := &fakeServer{respErr: status.Error(codes.Internal, "boom")}
	client, cleanup := startTestServer(t, fake)
	t.Cleanup(cleanup)

	uploader := NewGrpcUploader(client)
	_, err := uploader.Upload(context.Background(), strings.NewReader("x"), ".png")
	require.Error(t, err)
	require.Contains(t, err.Error(), "upload rpc")
}

type errReader struct{}

func (errReader) Read([]byte) (int, error) { return 0, errors.New("read failed") }

func TestGrpcUploader_ReadErrorIsWrapped(t *testing.T) {
	t.Parallel()

	fake := &fakeServer{}
	client, cleanup := startTestServer(t, fake)
	t.Cleanup(cleanup)

	uploader := NewGrpcUploader(client)
	_, err := uploader.Upload(context.Background(), io.MultiReader(errReader{}), ".png")
	require.Error(t, err)
	require.Contains(t, err.Error(), "read body")
}

func TestGrpcUploader_RespectsContextDeadline(t *testing.T) {
	t.Parallel()

	fake := &fakeServer{respErr: status.Error(codes.DeadlineExceeded, "slow")}
	client, cleanup := startTestServer(t, fake)
	t.Cleanup(cleanup)

	ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
	defer cancel()

	uploader := NewGrpcUploader(client)
	_, err := uploader.Upload(ctx, strings.NewReader("x"), ".png")
	require.Error(t, err)
}
