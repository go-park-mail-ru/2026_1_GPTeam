package validators_test

import (
	"testing"
	"time"

	validators2 "github.com/go-park-mail-ru/2026_1_GPTeam/pkg/validators"
	"github.com/stretchr/testify/require"
)

func TestValidateUsername(t *testing.T) {
	testCases := []struct {
		Name     string
		Username string
		err      error
	}{
		{"Short length", "ab", validators2.UsernameShortError},
		{"Correct", "admin123", nil},
		{"Incorrect", "admi_n", validators2.UsernameWrongSymbolsError},
	}

	for _, testCase := range testCases {
		t.Run(testCase.Name, func(t *testing.T) {
			t.Parallel()
			err := validators2.ValidateUsername(testCase.Username)
			require.ErrorIs(t, err, testCase.err)
		})
	}
}

func TestValidatePassword(t *testing.T) {
	testCases := []struct {
		Name     string
		Password string
		err      error
	}{
		{"Short length", "ab", validators2.IncorrectPasswordError},
		{"Has no upper", "admin123", validators2.IncorrectPasswordError},
		{"Has no lower", "ADMIN123", validators2.IncorrectPasswordError},
		{"Has no digit", "AdminAdmin", validators2.IncorrectPasswordError},
		{"Has invalid symbols", "Admin123!", validators2.IncorrectPasswordError},
		{"Correct", "Admin123", nil},
	}

	for _, testCase := range testCases {
		t.Run(testCase.Name, func(t *testing.T) {
			t.Parallel()
			err := validators2.ValidatePassword(testCase.Password)
			require.ErrorIs(t, err, testCase.err)
		})
	}
}

func TestValidateEmail(t *testing.T) {
	testCases := []struct {
		Name  string
		Email string
		err   error
	}{
		{"Incorrect length", "", validators2.EmailError},
		{"Incorrect email", "abc", validators2.EmailError},
		{"Incorrect email", "abc@", validators2.EmailError},
		{"Incorrect email", "abc@abc", validators2.EmailError},
		{"Incorrect email", "abc@abc.1", validators2.EmailError},
		{"Correct email", "abc@example.com", nil},
	}

	for _, testCase := range testCases {
		t.Run(testCase.Name, func(t *testing.T) {
			t.Parallel()
			err := validators2.ValidateEmail(testCase.Email)
			require.ErrorIs(t, err, testCase.err)
		})
	}
}

func TestValidateCurrency(t *testing.T) {
	testCases := []struct {
		Name     string
		Currency string
		err      error
	}{
		{"Incorrect", "abc", validators2.CurrencyNotAllowed},
		{"Incorrect", "ABC", validators2.CurrencyNotAllowed},
		{"Incorrect", "GBP", validators2.CurrencyNotAllowed},
		{"Correct", "RUB", nil},
		{"Correct", "USD", nil},
		{"Correct", "EUR", nil},
	}

	for _, testCase := range testCases {
		t.Run(testCase.Name, func(t *testing.T) {
			t.Parallel()
			err := validators2.ValidateCurrency(testCase.Currency, []string{
				"RUB", "USD", "EUR",
			})
			require.ErrorIs(t, err, testCase.err)
		})
	}
}

func TestValidateTargetBudget(t *testing.T) {
	testCases := []struct {
		Name   string
		Target int
		err    error
	}{
		{"Negative", -1, validators2.TargetIsNegativeError},
		{"Big", 1e18 + 1, validators2.TargetIsBigError},
		{"Correct", 1000, nil},
	}

	for _, testCase := range testCases {
		t.Run(testCase.Name, func(t *testing.T) {
			t.Parallel()
			err := validators2.ValidateTargetBudget(testCase.Target)
			require.ErrorIs(t, err, testCase.err)
		})
	}
}

func TestValidateStartDate(t *testing.T) {
	testCases := []struct {
		Name  string
		Start time.Time
		err   error
	}{
		{"In past", time.Now().AddDate(0, 0, -1), validators2.StartDateInPastError},
		{"Correct", time.Now().AddDate(0, 0, 1), nil},
	}

	for _, testCase := range testCases {
		t.Run(testCase.Name, func(t *testing.T) {
			t.Parallel()
			err := validators2.ValidateStartDate(testCase.Start)
			require.ErrorIs(t, err, testCase.err)
		})
	}
}

func TestValidateEndDate(t *testing.T) {
	testCases := []struct {
		Name  string
		Start time.Time
		End   time.Time
		err   error
	}{
		{"In past", time.Now(), time.Now().AddDate(0, 0, -1), validators2.EndDateInPastError},
		{"Correct (in future)", time.Now(), time.Now().AddDate(0, 0, 1), nil},
		{"Correct (nil end time)", time.Now(), time.Time{}, nil},
	}

	for _, testCase := range testCases {
		t.Run(testCase.Name, func(t *testing.T) {
			t.Parallel()
			err := validators2.ValidateEndDate(testCase.Start, testCase.End)
			require.ErrorIs(t, err, testCase.err)
		})
	}
}
