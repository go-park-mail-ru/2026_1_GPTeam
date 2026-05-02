package currency_converter_test

import (
	"strconv"
	"testing"

	"github.com/go-park-mail-ru/2026_1_GPTeam/pkg/currency_converter"
	"github.com/stretchr/testify/require"
)

func TestConvertToRub(t *testing.T) {
	testCases := []struct {
		CurrencyFrom string
		Value        float64
		Result       float64
	}{
		{"EUR", 1, 88},
		{"USD", 1, 75},
		{"RUB", 1, 1},
		{"AAA", 1, 1},
	}
	for i, testCase := range testCases {
		t.Run(strconv.Itoa(i), func(t *testing.T) {
			t.Parallel()
			res := currency_converter.ConvertToRub(testCase.Value, testCase.CurrencyFrom)
			require.Equal(t, testCase.Result, res)
		})
	}
}
