package validators_test

import (
	"testing"
	"time"

	"github.com/go-park-mail-ru/2026_1_GPTeam/internal/web/web_helpers"
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
		{"Incorrect email without at", "abc", validators2.EmailError},
		{"Incorrect email with at", "abc@", validators2.EmailError},
		{"Incorrect email without dot", "abc@abc", validators2.EmailError},
		{"Incorrect email with digit tld", "abc@abc.1", validators2.EmailError},
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
		{"Incorrect lowercase", "abc", validators2.CurrencyNotAllowedError},
		{"Incorrect uppercase not in list", "ABC", validators2.CurrencyNotAllowedError},
		{"Incorrect GBP not allowed", "GBP", validators2.CurrencyNotAllowedError},
		{"Correct RUB", "RUB", nil},
		{"Correct USD", "USD", nil},
		{"Correct EUR", "EUR", nil},
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
		{"Zero", 0, validators2.TargetIsZeroError},
		{"Big", 1_000_000_000_001, validators2.TargetIsBigError},
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
			err := validators2.ValidateBudgetStartDate(testCase.Start)
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
			err := validators2.ValidateBudgetEndDate(testCase.Start, testCase.End)
			require.ErrorIs(t, err, testCase.err)
		})
	}
}

func TestValidateTransaction(t *testing.T) {
	allowed := []string{"income", "expense"}
	categories := []string{"food", "salary"}

	testCases := []struct {
		Name   string
		Body   web_helpers.TransactionRequest
		errLen int
	}{
		{
			"All correct",
			web_helpers.TransactionRequest{
				Title:       "Test",
				Description: "Desc",
				Value:       100,
				Type:        "income",
				Category:    "salary",
			},
			0,
		},
		{
			"Empty title",
			web_helpers.TransactionRequest{
				Title:       "",
				Description: "Desc",
				Value:       100,
				Type:        "income",
				Category:    "salary",
			},
			1,
		},
		{
			"Negative value",
			web_helpers.TransactionRequest{
				Title:       "Test",
				Description: "Desc",
				Value:       -1,
				Type:        "income",
				Category:    "salary",
			},
			1,
		},
		{
			"Invalid type and category",
			web_helpers.TransactionRequest{
				Title:       "Test",
				Description: "Desc",
				Value:       100,
				Type:        "wrong",
				Category:    "wrong",
			},
			2,
		},
		{
			"All wrong",
			web_helpers.TransactionRequest{
				Title:       "",
				Description: "",
				Value:       -1,
				Type:        "wrong",
				Category:    "wrong",
			},
			5,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.Name, func(t *testing.T) {
			t.Parallel()
			errs := validators2.ValidateTransaction(testCase.Body, allowed, categories)
			require.Len(t, errs, testCase.errLen)
		})
	}
}
