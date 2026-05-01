package validators_test

import (
	"strings"
	"testing"
	"time"

	"github.com/go-park-mail-ru/2026_1_GPTeam/internal/web/web_helpers"
	"github.com/go-park-mail-ru/2026_1_GPTeam/pkg/validators"
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
		{"Incorrect symbols", "admi_n", validators.UsernameWrongSymbolsError},
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

func TestValidateConfirmPassword(t *testing.T) {
	testCases := []struct {
		Name            string
		Password        string
		ConfirmPassword string
		err             error
	}{
		{"Passwords match", "Admin123", "Admin123", nil},
		{"Passwords do not match", "Admin123", "Admin456", validators.PasswordsNotSameError},
	}

	for _, testCase := range testCases {
		t.Run(testCase.Name, func(t *testing.T) {
			t.Parallel()
			err := validators.ValidateConfirmPassword(testCase.Password, testCase.ConfirmPassword)
			require.ErrorIs(t, err, testCase.err)
		})
	}
}

func TestValidateEmail(t *testing.T) {
	longEmail := strings.Repeat("a", 245) + "@example.com"
	testCases := []struct {
		Name  string
		Email string
		err   error
	}{
		{"Empty email", "", validators.EmailError},
		{"Too long email", longEmail, validators.EmailError},
		{"Incorrect email without at", "abc", validators.EmailError},
		{"Incorrect email with at", "abc@", validators.EmailError},
		{"Incorrect email without dot", "abc@abc", validators.EmailError},
		{"Incorrect email with digit tld", "abc@abc.1", validators.EmailError},
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

func TestValidateBudgetTitle(t *testing.T) {
	testCases := []struct {
		Name  string
		Title string
		err   error
	}{
		{"Empty title", "", validators.BudgetTitleEmpty},
		{"Too long title", strings.Repeat("a", 256), validators.BudgetTitleTooLong},
		{"Correct title", "My Budget", nil},
	}

	for _, testCase := range testCases {
		t.Run(testCase.Name, func(t *testing.T) {
			t.Parallel()
			err := validators.ValidateBudgetTitle(testCase.Title)
			require.ErrorIs(t, err, testCase.err)
		})
	}
}

func TestValidateBudgetDescription(t *testing.T) {
	testCases := []struct {
		Name        string
		Description string
		err         error
	}{
		{"Empty description", "", validators.BudgetDescriptionEmpty},
		{"Correct description", "For vacation", nil},
	}

	for _, testCase := range testCases {
		t.Run(testCase.Name, func(t *testing.T) {
			t.Parallel()
			err := validators.ValidateBudgetDescription(testCase.Description)
			require.ErrorIs(t, err, testCase.err)
		})
	}
}

func TestValidateCurrency(t *testing.T) {
	allowed := []string{"RUB", "USD", "EUR"}
	testCases := []struct {
		Name     string
		Currency string
		err      error
	}{
		{"Incorrect lowercase", "abc", validators.CurrencyNotAllowedError},
		{"Incorrect uppercase not in list", "ABC", validators.CurrencyNotAllowedError},
		{"Incorrect GBP not allowed", "GBP", validators.CurrencyNotAllowedError},
		{"Correct RUB", "RUB", nil},
		{"Correct USD", "USD", nil},
		{"Correct EUR", "EUR", nil},
	}

	for _, testCase := range testCases {
		t.Run(testCase.Name, func(t *testing.T) {
			t.Parallel()
			err := validators.ValidateCurrency(testCase.Currency, allowed)
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
		{"Zero", 0, validators.TargetIsZeroError},
		{"Big", 1_000_000_001, validators.TargetIsBigError},
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

func TestValidateActualBudget(t *testing.T) {
	testCases := []struct {
		Name   string
		Actual int
		err    error
	}{
		{"Negative", -1, validators.ValueIsNegativeError},
		{"Big", 1_000_000_001, validators.ValueIsBigError},
		{"Zero is allowed", 0, nil},
		{"Correct", 500, nil},
	}

	for _, testCase := range testCases {
		t.Run(testCase.Name, func(t *testing.T) {
			t.Parallel()
			err := validators.ValidateActualBudget(testCase.Actual)
			require.ErrorIs(t, err, testCase.err)
		})
	}
}

func TestValidateBudgetStartDate(t *testing.T) {
	testCases := []struct {
		Name  string
		Start time.Time
		err   error
	}{
		{"In past", time.Now().AddDate(0, 0, -1), validators.StartDateInPastError},
		{"Correct future", time.Now().AddDate(0, 0, 1), nil},
		{"Correct today", time.Now(), nil},
	}

	for _, testCase := range testCases {
		t.Run(testCase.Name, func(t *testing.T) {
			t.Parallel()
			err := validators.ValidateBudgetStartDate(testCase.Start)
			require.ErrorIs(t, err, testCase.err)
		})
	}
}

func TestValidateBudgetEndDate(t *testing.T) {
	testCases := []struct {
		Name  string
		Start time.Time
		End   time.Time
		err   error
	}{
		{"End before start", time.Now(), time.Now().AddDate(0, 0, -1), validators.EndDateInPastError},
		{"Correct (in future)", time.Now(), time.Now().AddDate(0, 0, 1), nil},
		{"Correct (nil end time)", time.Now(), time.Time{}, nil},
	}

	for _, testCase := range testCases {
		t.Run(testCase.Name, func(t *testing.T) {
			t.Parallel()
			err := validators.ValidateBudgetEndDate(testCase.Start, testCase.End)
			require.ErrorIs(t, err, testCase.err)
		})
	}
}

func TestValidateBudget(t *testing.T) {
	t.Parallel()
	currencies := []string{"RUB", "USD"}

	t.Run("All correct", func(t *testing.T) {
		body := web_helpers.BudgetRequest{
			Title:       "Vacation",
			Description: "Savings",
			Currency:    "USD",
			Target:      1000,
			Actual:      500,
			StartAt:     time.Now().AddDate(0, 0, 1),
			EndAt:       time.Now().AddDate(0, 1, 0),
		}
		errs := validators.ValidateBudget(body, currencies)
		require.Empty(t, errs)
	})

	t.Run("Multiple errors", func(t *testing.T) {
		body := web_helpers.BudgetRequest{
			Title:       "",
			Description: "",
			Currency:    "GBP",
			Target:      -1,
			Actual:      -1,
			StartAt:     time.Now().AddDate(0, 0, -2),
			EndAt:       time.Now().AddDate(0, 0, -5),
		}
		errs := validators.ValidateBudget(body, currencies)
		require.Len(t, errs, 7)
	})
}

func TestValidateTransactionTitle(t *testing.T) {
	testCases := []struct {
		Name  string
		Title string
		err   error
	}{
		{"Empty title", "   ", validators.TransactionTitleEmptyError},
		{"Too long title", strings.Repeat("а", 256), validators.TransactionTitleLongError},
		{"Correct title", "Groceries", nil},
	}

	for _, testCase := range testCases {
		t.Run(testCase.Name, func(t *testing.T) {
			t.Parallel()
			err := validators.ValidateTransactionTitle(testCase.Title)
			require.ErrorIs(t, err, testCase.err)
		})
	}
}

func TestValidateTransactionDescription(t *testing.T) {
	testCases := []struct {
		Name        string
		Description string
		err         error
	}{
		{"Empty description", "   ", validators.TransactionDescriptionEmptyError},
		{"Correct description", "Weekly shopping", nil},
	}

	for _, testCase := range testCases {
		t.Run(testCase.Name, func(t *testing.T) {
			t.Parallel()
			err := validators.ValidateTransactionDescription(testCase.Description)
			require.ErrorIs(t, err, testCase.err)
		})
	}
}

func TestValidateTransactionValue(t *testing.T) {
	testCases := []struct {
		Name  string
		Value float64
		err   error
	}{
		{"Negative", -10.5, validators.ValueIsNegativeError},
		{"Zero", 0, validators.ValueIsNegativeError},
		{"Too big", 1_000_000_001, validators.ValueIsBigError},
		{"Correct", 100.5, nil},
	}

	for _, testCase := range testCases {
		t.Run(testCase.Name, func(t *testing.T) {
			t.Parallel()
			err := validators.ValidateTransactionValue(testCase.Value)
			require.ErrorIs(t, err, testCase.err)
		})
	}
}

func TestValidateTransactionType(t *testing.T) {
	allowed := []string{"income", "expense"}
	testCases := []struct {
		Name string
		Type string
		err  error
	}{
		{"Invalid type", "transfer", validators.TransactionTypeNotAllowedError},
		{"Valid type", "income", nil},
	}

	for _, testCase := range testCases {
		t.Run(testCase.Name, func(t *testing.T) {
			t.Parallel()
			err := validators.ValidateTransactionType(testCase.Type, allowed)
			require.ErrorIs(t, err, testCase.err)
		})
	}
}

func TestValidateTransactionCategory(t *testing.T) {
	allowed := []string{"food", "salary"}
	testCases := []struct {
		Name     string
		Category string
		err      error
	}{
		{"Invalid category", "entertainment", validators.TransactionCategoryNotAllowedError},
		{"Valid category", "food", nil},
	}

	for _, testCase := range testCases {
		t.Run(testCase.Name, func(t *testing.T) {
			t.Parallel()
			err := validators.ValidateTransactionCategory(testCase.Category, allowed)
			require.ErrorIs(t, err, testCase.err)
		})
	}
}

func TestValidateTransaction(t *testing.T) {
	allowedTypes := []string{"income", "expense"}
	categories := []string{"food", "salary"}
	currencies := []string{"RUB", "USD", "EUR"}

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
				Currency:    "RUB",
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
				Currency:    "RUB",
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
				Currency:    "RUB",
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
				Currency:    "RUB",
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
				Currency:    "GBP",
			},
			5,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.Name, func(t *testing.T) {
			t.Parallel()
			errs := validators.ValidateTransaction(testCase.Body, allowedTypes, categories, currencies)
			require.Len(t, errs, testCase.errLen)
		})
	}
}

func TestValidateSignUpUser(t *testing.T) {
	t.Parallel()

	t.Run("Valid sign up", func(t *testing.T) {
		body := web_helpers.SignupBodyRequest{
			Username:        "admin123",
			Password:        "Admin123",
			ConfirmPassword: "Admin123",
			Email:           "test@example.com",
		}
		errs := validators.ValidateSignUpUser(body)
		require.Empty(t, errs)
	})

	t.Run("Multiple errors in sign up", func(t *testing.T) {
		body := web_helpers.SignupBodyRequest{
			Username:        "ab",
			Password:        "admin",
			ConfirmPassword: "wrong",
			Email:           "invalid",
		}
		errs := validators.ValidateSignUpUser(body)
		require.Len(t, errs, 5)
	})
}

func TestValidateUpdateUser(t *testing.T) {
	t.Parallel()

	ptr := func(s string) *string { return &s }

	t.Run("No fields to update", func(t *testing.T) {
		body := web_helpers.UpdateUserProfileRequest{}
		errs := validators.ValidateUpdateUser(body)
		require.Len(t, errs, 1)
		require.Equal(t, "request", errs[0].Field)
	})

	t.Run("All fields valid", func(t *testing.T) {
		body := web_helpers.UpdateUserProfileRequest{
			Username: ptr("newadmin"),
			Email:    ptr("new@example.com"),
			Password: ptr("NewAdmin123"),
		}
		errs := validators.ValidateUpdateUser(body)
		require.Empty(t, errs)
	})

	t.Run("All fields invalid", func(t *testing.T) {
		body := web_helpers.UpdateUserProfileRequest{
			Username: ptr("ab"),
			Email:    ptr("invalid-email"),
			Password: ptr("weak"),
		}
		errs := validators.ValidateUpdateUser(body)
		require.Len(t, errs, 3)
	})
}
