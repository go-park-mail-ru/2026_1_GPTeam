package storage

import (
	"context"
	"errors"
	"io"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestLocalStorage_SaveSuccess(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	s := NewLocalStorage(dir)

	name, err := s.Save(context.Background(), strings.NewReader("hello"), ".txt")
	require.NoError(t, err)
	require.NotEmpty(t, name)
	require.True(t, strings.HasSuffix(name, ".txt"))

	got, err := os.ReadFile(filepath.Join(dir, name))
	require.NoError(t, err)
	require.Equal(t, "hello", string(got))
}

func TestLocalStorage_CreatesDirectoryWhenMissing(t *testing.T) {
	t.Parallel()

	parent := t.TempDir()
	dir := filepath.Join(parent, "nested", "static")
	s := NewLocalStorage(dir)

	name, err := s.Save(context.Background(), strings.NewReader("data"), ".bin")
	require.NoError(t, err)

	_, err = os.Stat(filepath.Join(dir, name))
	require.NoError(t, err)
}

func TestLocalStorage_GeneratesUniqueNames(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	s := NewLocalStorage(dir)

	name1, err := s.Save(context.Background(), strings.NewReader("a"), ".png")
	require.NoError(t, err)
	name2, err := s.Save(context.Background(), strings.NewReader("b"), ".png")
	require.NoError(t, err)

	require.NotEqual(t, name1, name2)
}

func TestLocalStorage_RespectsCancelledContext(t *testing.T) {
	t.Parallel()

	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	s := NewLocalStorage(t.TempDir())
	_, err := s.Save(ctx, strings.NewReader("x"), ".txt")
	require.ErrorIs(t, err, context.Canceled)
}

type errReader struct{}

func (errReader) Read([]byte) (int, error) { return 0, errors.New("boom") }

func TestLocalStorage_RemovesPartialFileOnReadError(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	s := NewLocalStorage(dir)

	_, err := s.Save(context.Background(), io.MultiReader(errReader{}), ".bin")
	require.Error(t, err)

	entries, err := os.ReadDir(dir)
	require.NoError(t, err)
	require.Empty(t, entries, "partially written file must be removed")
}
