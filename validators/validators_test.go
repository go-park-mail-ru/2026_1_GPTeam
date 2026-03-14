package validators_test

import (
	"testing"
	"time"

	"github.com/go-park-mail-ru/2026_1_GPTeam/validators"
	"github.com/stretchr/testify/require"
)

func TestValidateUsername(t *testing.T) {
	testCases := []struct {
		Name     string
		Username string
		err      error
	}{
		{"Short length", "ab", validators.UsernameShortError},
		{"Correct", "admin123", nil},
		{"Incorrect", "admi_n", validators.UsernameWrongSymbolsError},
	}

	for _, testCase := range testCases {
		t.Run(testCase.Name, func(t *testing.T) {
			t.Parallel()
			err := validators.ValidateUsername(testCase.Username)
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
		{"Short length", "ab", validators.IncorrectPasswordError},
		{"Has no upper", "admin123", validators.IncorrectPasswordError},
		{"Has no lower", "ADMIN123", validators.IncorrectPasswordError},
		{"Has no digit", "AdminAdmin", validators.IncorrectPasswordError},
		{"Has invalid symbols", "Admin123!", validators.IncorrectPasswordError},
		{"Correct", "Admin123", nil},
	}

	for _, testCase := range testCases {
		t.Run(testCase.Name, func(t *testing.T) {
			t.Parallel()
			err := validators.ValidatePassword(testCase.Password)
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
		{"Incorrect length", "", validators.EmailError},
		{"Incorrect email", "abc", validators.EmailError},
		{"Incorrect email", "abc@", validators.EmailError},
		{"Incorrect email", "abc@abc", validators.EmailError},
		{"Incorrect email", "abc@abc.1", validators.EmailError},
		{"Correct email", "abc@example.com", nil},
	}

	for _, testCase := range testCases {
		t.Run(testCase.Name, func(t *testing.T) {
			t.Parallel()
			err := validators.ValidateEmail(testCase.Email)
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
		{"Incorrect", "abc", validators.CurrencyNotAllowed},
		{"Incorrect", "ABC", validators.CurrencyNotAllowed},
		{"Incorrect", "GBP", validators.CurrencyNotAllowed},
		{"Correct", "RUB", nil},
		{"Correct", "USD", nil},
		{"Correct", "EUR", nil},
	}

	for _, testCase := range testCases {
		t.Run(testCase.Name, func(t *testing.T) {
			t.Parallel()
			err := validators.ValidateCurrency(testCase.Currency)
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
		{"Negative", -1, validators.TargetIsNegativeError},
		{"Big", 1e18 + 1, validators.TargetIsBigError},
		{"Correct", 1000, nil},
	}

	for _, testCase := range testCases {
		t.Run(testCase.Name, func(t *testing.T) {
			t.Parallel()
			err := validators.ValidateTargetBudget(testCase.Target)
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
		{"In past", time.Now().AddDate(0, 0, -1), validators.StartDateInPastError},
		{"Correct", time.Now().AddDate(0, 0, 1), nil},
	}

	for _, testCase := range testCases {
		t.Run(testCase.Name, func(t *testing.T) {
			t.Parallel()
			err := validators.ValidateStartDate(testCase.Start)
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
		{"In past", time.Now(), time.Now().AddDate(0, 0, -1), validators.EndDateInPastError},
		{"Correct (in future)", time.Now(), time.Now().AddDate(0, 0, 1), nil},
		{"Correct (nil end time)", time.Now(), time.Time{}, nil},
	}

	for _, testCase := range testCases {
		t.Run(testCase.Name, func(t *testing.T) {
			t.Parallel()
			err := validators.ValidateEndDate(testCase.Start, testCase.End)
			require.ErrorIs(t, err, testCase.err)
		})
	}
}
