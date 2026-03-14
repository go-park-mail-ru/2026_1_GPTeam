package testhelper

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/require"
)

func MustJSON(t *testing.T, value any) []byte {
	t.Helper()
	bytes, err := json.Marshal(value)
	require.NoError(t, err)
	return bytes
}
