package auth

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	authv1 "github.com/go-park-mail-ru/2026_1_GPTeam/pkg/gen/auth/v1"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc"
)

// fakeAuthServiceClient — минимальный in-memory фейк для тестов GrpcAuthAdapter.
type fakeAuthServiceClient struct {
	issueResp *authv1.IssueTokensResponse
	issueErr  error

	refreshResp *authv1.RefreshResponse
	refreshErr  error

	revokeResp *authv1.RevokeResponse
	revokeErr  error

	gotIssueReq   *authv1.IssueTokensRequest
	gotRefreshReq *authv1.RefreshRequest
	gotRevokeReq  *authv1.RevokeRequest
}

func (f *fakeAuthServiceClient) IssueTokens(ctx context.Context, in *authv1.IssueTokensRequest, opts ...grpc.CallOption) (*authv1.IssueTokensResponse, error) {
	f.gotIssueReq = in
	return f.issueResp, f.issueErr
}

func (f *fakeAuthServiceClient) ValidateAccess(ctx context.Context, in *authv1.ValidateAccessRequest, opts ...grpc.CallOption) (*authv1.ValidateAccessResponse, error) {
	return nil, nil
}

func (f *fakeAuthServiceClient) Refresh(ctx context.Context, in *authv1.RefreshRequest, opts ...grpc.CallOption) (*authv1.RefreshResponse, error) {
	f.gotRefreshReq = in
	return f.refreshResp, f.refreshErr
}

func (f *fakeAuthServiceClient) Revoke(ctx context.Context, in *authv1.RevokeRequest, opts ...grpc.CallOption) (*authv1.RevokeResponse, error) {
	f.gotRevokeReq = in
	return f.revokeResp, f.revokeErr
}

func TestNewGrpcAuthAdapter(t *testing.T) {
	t.Parallel()

	fake := &fakeAuthServiceClient{}
	a := NewGrpcAuthAdapter(fake, "secret-123", "v1")

	require.NotNil(t, a)
	require.Equal(t, []byte("secret-123"), a.secret)
	require.Equal(t, "v1", a.version)
}

func TestGrpcAuthAdapter_GenerateNewAuth(t *testing.T) {
	t.Parallel()

	t.Run("success: cookies set", func(t *testing.T) {
		t.Parallel()
		fake := &fakeAuthServiceClient{
			issueResp: &authv1.IssueTokensResponse{
				AccessToken:  "access-tok",
				RefreshToken: "refresh-tok",
			},
		}
		a := NewGrpcAuthAdapter(fake, "secret", "v1")
		w := httptest.NewRecorder()

		a.GenerateNewAuth(context.Background(), w, 42)

		require.NotNil(t, fake.gotIssueReq)
		require.Equal(t, int32(42), fake.gotIssueReq.GetUserId())
		require.Equal(t, defaultDeviceID, fake.gotIssueReq.GetDeviceId())

		cookies := w.Result().Cookies()
		require.Len(t, cookies, 2)
		require.Equal(t, TokenName, cookies[0].Name)
		require.Equal(t, "access-tok", cookies[0].Value)
		require.Equal(t, RefreshTokenName, cookies[1].Name)
		require.Equal(t, "refresh-tok", cookies[1].Value)
	})

	t.Run("grpc error: no cookies", func(t *testing.T) {
		t.Parallel()
		fake := &fakeAuthServiceClient{
			issueErr: errors.New("boom"),
		}
		a := NewGrpcAuthAdapter(fake, "secret", "v1")
		w := httptest.NewRecorder()

		a.GenerateNewAuth(context.Background(), w, 7)

		require.Empty(t, w.Result().Cookies())
	})
}

func TestGrpcAuthAdapter_IsAuth(t *testing.T) {
	t.Parallel()

	t.Run("no cookie", func(t *testing.T) {
		t.Parallel()
		a := NewGrpcAuthAdapter(&fakeAuthServiceClient{}, "secret", "v1")

		req := httptest.NewRequest(http.MethodGet, "/", nil)
		ok, id := a.IsAuth(context.Background(), req)

		require.False(t, ok)
		require.Equal(t, -1, id)
	})

	t.Run("invalid token in cookie", func(t *testing.T) {
		t.Parallel()
		a := NewGrpcAuthAdapter(&fakeAuthServiceClient{}, "secret", "v1")

		req := httptest.NewRequest(http.MethodGet, "/", nil)
		req.AddCookie(&http.Cookie{Name: TokenName, Value: "garbage-token"})

		ok, id := a.IsAuth(context.Background(), req)

		require.False(t, ok)
		require.Equal(t, -1, id)
	})
}

func TestGrpcAuthAdapter_ClearOld(t *testing.T) {
	t.Parallel()

	t.Run("no refresh cookie: revoke not called, cookies cleared", func(t *testing.T) {
		t.Parallel()
		fake := &fakeAuthServiceClient{}
		a := NewGrpcAuthAdapter(fake, "secret", "v1")
		w := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodPost, "/logout", nil)

		a.ClearOld(context.Background(), w, req)

		require.Nil(t, fake.gotRevokeReq)
		require.Len(t, w.Result().Cookies(), 2)
	})

	t.Run("with refresh cookie: Revoke called, cookies cleared", func(t *testing.T) {
		t.Parallel()
		fake := &fakeAuthServiceClient{
			revokeResp: &authv1.RevokeResponse{},
		}
		a := NewGrpcAuthAdapter(fake, "secret", "v1")
		w := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodPost, "/logout", nil)
		req.AddCookie(&http.Cookie{Name: RefreshTokenName, Value: "old-refresh"})

		a.ClearOld(context.Background(), w, req)

		require.NotNil(t, fake.gotRevokeReq)
		require.Equal(t, "old-refresh", fake.gotRevokeReq.GetRefreshToken())
		require.Len(t, w.Result().Cookies(), 2)
	})

	t.Run("with refresh cookie and revoke error: still clears cookies", func(t *testing.T) {
		t.Parallel()
		fake := &fakeAuthServiceClient{
			revokeErr: errors.New("boom"),
		}
		a := NewGrpcAuthAdapter(fake, "secret", "v1")
		w := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodPost, "/logout", nil)
		req.AddCookie(&http.Cookie{Name: RefreshTokenName, Value: "x"})

		a.ClearOld(context.Background(), w, req)

		require.NotNil(t, fake.gotRevokeReq)
		require.Len(t, w.Result().Cookies(), 2)
	})
}

func TestGrpcAuthAdapter_Refresh(t *testing.T) {
	t.Parallel()

	t.Run("no refresh cookie", func(t *testing.T) {
		t.Parallel()
		a := NewGrpcAuthAdapter(&fakeAuthServiceClient{}, "secret", "v1")
		w := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodPost, "/refresh", nil)

		ok, id := a.Refresh(context.Background(), w, req)

		require.False(t, ok)
		require.Equal(t, -1, id)
		require.Empty(t, w.Result().Cookies())
	})

	t.Run("grpc error", func(t *testing.T) {
		t.Parallel()
		fake := &fakeAuthServiceClient{
			refreshErr: errors.New("boom"),
		}
		a := NewGrpcAuthAdapter(fake, "secret", "v1")
		w := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodPost, "/refresh", nil)
		req.AddCookie(&http.Cookie{Name: RefreshTokenName, Value: "rt"})

		ok, id := a.Refresh(context.Background(), w, req)

		require.False(t, ok)
		require.Equal(t, -1, id)
		require.Empty(t, w.Result().Cookies())
	})

	t.Run("response invalid", func(t *testing.T) {
		t.Parallel()
		fake := &fakeAuthServiceClient{
			refreshResp: &authv1.RefreshResponse{Valid: false},
		}
		a := NewGrpcAuthAdapter(fake, "secret", "v1")
		w := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodPost, "/refresh", nil)
		req.AddCookie(&http.Cookie{Name: RefreshTokenName, Value: "rt"})

		ok, id := a.Refresh(context.Background(), w, req)

		require.False(t, ok)
		require.Equal(t, -1, id)
		require.Empty(t, w.Result().Cookies())
	})

	t.Run("response valid: cookies set", func(t *testing.T) {
		t.Parallel()
		fake := &fakeAuthServiceClient{
			refreshResp: &authv1.RefreshResponse{
				Valid:        true,
				UserId:       77,
				AccessToken:  "new-access",
				RefreshToken: "new-refresh",
			},
		}
		a := NewGrpcAuthAdapter(fake, "secret", "v1")
		w := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodPost, "/refresh", nil)
		req.AddCookie(&http.Cookie{Name: RefreshTokenName, Value: "rt"})

		ok, id := a.Refresh(context.Background(), w, req)

		require.True(t, ok)
		require.Equal(t, 77, id)

		require.NotNil(t, fake.gotRefreshReq)
		require.Equal(t, "rt", fake.gotRefreshReq.GetRefreshToken())
		require.Equal(t, defaultDeviceID, fake.gotRefreshReq.GetDeviceId())

		cookies := w.Result().Cookies()
		require.Len(t, cookies, 2)
		require.Equal(t, "new-access", cookies[0].Value)
		require.Equal(t, "new-refresh", cookies[1].Value)
	})
}
