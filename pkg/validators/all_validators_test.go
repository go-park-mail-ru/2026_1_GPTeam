package validators

import (
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func TestValidateUsername(t *testing.T) {
	testCases := []struct {
		Name     string
		Username string
		err      error
	}{
		{"Short length", "ab", UsernameShortError},
		{"Correct", "admin123", nil},
		{"Incorrect", "admi_n", UsernameWrongSymbolsError},
	}

	for _, testCase := range testCases {
		t.Run(testCase.Name, func(t *testing.T) {
			t.Parallel()
			err := validateUsername(testCase.Username)
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
		{"Short length", "ab", IncorrectPasswordError},
		{"Has no upper", "admin123", IncorrectPasswordError},
		{"Has no lower", "ADMIN123", IncorrectPasswordError},
		{"Has no digit", "AdminAdmin", IncorrectPasswordError},
		{"Has invalid symbols", "Admin123!", IncorrectPasswordError},
		{"Correct", "Admin123", nil},
	}

	for _, testCase := range testCases {
		t.Run(testCase.Name, func(t *testing.T) {
			t.Parallel()
			err := validatePassword(testCase.Password)
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
		{"Incorrect length", "", EmailError},
		{"Incorrect email", "abc", EmailError},
		{"Incorrect email", "abc@", EmailError},
		{"Incorrect email", "abc@abc", EmailError},
		{"Incorrect email", "abc@abc.1", EmailError},
		{"Correct email", "abc@example.com", nil},
	}

	for _, testCase := range testCases {
		t.Run(testCase.Name, func(t *testing.T) {
			t.Parallel()
			err := validateEmail(testCase.Email)
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
		{"Incorrect", "abc", CurrencyNotAllowedError},
		{"Incorrect", "ABC", CurrencyNotAllowedError},
		{"Incorrect", "GBP", CurrencyNotAllowedError},
		{"Correct", "RUB", nil},
		{"Correct", "USD", nil},
		{"Correct", "EUR", nil},
	}

	for _, testCase := range testCases {
		t.Run(testCase.Name, func(t *testing.T) {
			t.Parallel()
			err := validateCurrency(testCase.Currency, []string{
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
		{"Negative", -1, TargetIsNegativeError},
		{"Big", 1e18 + 1, ValueIsBigError},
		{"Correct", 1000, nil},
	}

	for _, testCase := range testCases {
		t.Run(testCase.Name, func(t *testing.T) {
			t.Parallel()
			err := validateTargetBudget(testCase.Target)
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
		{"In past", time.Now().AddDate(0, 0, -1), StartDateInPastError},
		{"Correct", time.Now().AddDate(0, 0, 1), nil},
	}

	for _, testCase := range testCases {
		t.Run(testCase.Name, func(t *testing.T) {
			t.Parallel()
			err := validateBudgetStartDate(testCase.Start)
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
		{"In past", time.Now(), time.Now().AddDate(0, 0, -1), EndDateInPastError},
		{"Correct (in future)", time.Now(), time.Now().AddDate(0, 0, 1), nil},
		{"Correct (nil end time)", time.Now(), time.Time{}, nil},
	}

	for _, testCase := range testCases {
		t.Run(testCase.Name, func(t *testing.T) {
			t.Parallel()
			err := validateBudgetEndDate(testCase.Start, testCase.End)
			require.ErrorIs(t, err, testCase.err)
		})
	}
}
