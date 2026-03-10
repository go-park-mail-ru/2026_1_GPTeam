package testhelper

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/require"
)

func MustJSON(t *testing.T, v any) []byte {
	t.Helper()
	b, err := json.Marshal(v)
	require.NoError(t, err)
	return b
}
