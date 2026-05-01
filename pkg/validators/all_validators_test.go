package validators

import (
	"strings"
	"testing"
	"time"

	"github.com/go-park-mail-ru/2026_1_GPTeam/internal/web/web_helpers"
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
		{"Incorrect symbols", "admi_n", UsernameWrongSymbolsError},
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

func TestValidateConfirmPassword(t *testing.T) {
	testCases := []struct {
		Name            string
		Password        string
		ConfirmPassword string
		err             error
	}{
		{"Passwords match", "Admin123", "Admin123", nil},
		{"Passwords do not match", "Admin123", "Admin456", PasswordsNotSameError},
	}

	for _, testCase := range testCases {
		t.Run(testCase.Name, func(t *testing.T) {
			t.Parallel()
			err := validateConfirmPassword(testCase.Password, testCase.ConfirmPassword)
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
		{"Empty email", "", EmailError},
		{"Too long email", longEmail, EmailError},
		{"Incorrect email without at", "abc", EmailError},
		{"Incorrect email with at", "abc@", EmailError},
		{"Incorrect email without dot", "abc@abc", EmailError},
		{"Incorrect email with digit tld", "abc@abc.1", EmailError},
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

func TestValidateBudgetTitle(t *testing.T) {
	testCases := []struct {
		Name  string
		Title string
		err   error
	}{
		{"Empty title", "", BudgetTitleEmpty},
		{"Too long title", strings.Repeat("a", 256), BudgetTitleTooLong},
		{"Correct title", "My Budget", nil},
	}

	for _, testCase := range testCases {
		t.Run(testCase.Name, func(t *testing.T) {
			t.Parallel()
			err := validateBudgetTitle(testCase.Title)
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
		{"Empty description", "", BudgetDescriptionEmpty},
		{"Correct description", "For vacation", nil},
	}

	for _, testCase := range testCases {
		t.Run(testCase.Name, func(t *testing.T) {
			t.Parallel()
			err := validateBudgetDescription(testCase.Description)
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
		{"Incorrect lowercase", "abc", CurrencyNotAllowedError},
		{"Incorrect uppercase not in list", "ABC", CurrencyNotAllowedError},
		{"Incorrect GBP not allowed", "GBP", CurrencyNotAllowedError},
		{"Correct RUB", "RUB", nil},
		{"Correct USD", "USD", nil},
		{"Correct EUR", "EUR", nil},
	}

	for _, testCase := range testCases {
		t.Run(testCase.Name, func(t *testing.T) {
			t.Parallel()
			err := validateCurrency(testCase.Currency, allowed)
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
		{"Zero", 0, TargetIsZeroError},
		{"Big", 1_000_000_000_001, TargetIsBigError},
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

func TestValidateActualBudget(t *testing.T) {
	testCases := []struct {
		Name   string
		Actual int
		err    error
	}{
		{"Negative", -1, ValueIsNegativeError},
		{"Big", 1_000_000_000_001, ValueIsBigError},
		{"Zero is allowed", 0, nil},
		{"Correct", 500, nil},
	}

	for _, testCase := range testCases {
		t.Run(testCase.Name, func(t *testing.T) {
			t.Parallel()
			err := validateActualBudget(testCase.Actual)
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
		{"In past", time.Now().AddDate(0, 0, -1), StartDateInPastError},
		{"Correct future", time.Now().AddDate(0, 0, 1), nil},
		{"Correct today", time.Now(), nil},
	}

	for _, testCase := range testCases {
		t.Run(testCase.Name, func(t *testing.T) {
			t.Parallel()
			err := validateBudgetStartDate(testCase.Start)
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
		{"End before start", time.Now(), time.Now().AddDate(0, 0, -1), EndDateInPastError},
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

func TestValidateBudget(t *testing.T) {
	t.Parallel()
	currencies := []string{"RUB", "USD"}

	t.Run("All correct", func(t *testing.T) {
		body := web_helpers.BudgetRequest{
			Title:       "Vacation",
			Description: "Savings",
			Currency:    "RUB",
			Target:      1000,
			Actual:      500,
			StartAt:     time.Now().AddDate(0, 0, 1),
			EndAt:       time.Now().AddDate(0, 1, 0),
			Category:    []string{"Зарплата"},
		}
		allowedCategories := []string{
			"Зарплата",
			"Стипендия",
		}
		errs := ValidateBudget(body, currencies, allowedCategories)
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
			Category:    []string{},
		}
		allowedCategories := []string{
			"Зарплата",
			"Стипендия",
		}
		errs := ValidateBudget(body, currencies, allowedCategories)
		require.Len(t, errs, 8)
	})
}

func TestValidateTransactionTitle(t *testing.T) {
	testCases := []struct {
		Name  string
		Title string
		err   error
	}{
		{"Empty title", "   ", TransactionTitleEmptyError},
		{"Too long title", strings.Repeat("а", 256), TransactionTitleLongError},
		{"Correct title", "Groceries", nil},
	}

	for _, testCase := range testCases {
		t.Run(testCase.Name, func(t *testing.T) {
			t.Parallel()
			err := validateTransactionTitle(testCase.Title)
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
		{"Empty description", "   ", TransactionDescriptionEmptyError},
		{"Correct description", "Weekly shopping", nil},
	}

	for _, testCase := range testCases {
		t.Run(testCase.Name, func(t *testing.T) {
			t.Parallel()
			err := validateTransactionDescription(testCase.Description)
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
		{"Negative", -10.5, ValueIsNegativeError},
		{"Zero", 0, ValueIsNegativeError},
		{"Too big", 1_000_000_001, ValueIsBigError},
		{"Correct", 100.5, nil},
	}

	for _, testCase := range testCases {
		t.Run(testCase.Name, func(t *testing.T) {
			t.Parallel()
			err := validateTransactionValue(testCase.Value)
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
		{"Invalid type", "transfer", TransactionTypeNotAllowedError},
		{"Valid type", "income", nil},
	}

	for _, testCase := range testCases {
		t.Run(testCase.Name, func(t *testing.T) {
			t.Parallel()
			err := validateTransactionType(testCase.Type, allowed)
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
		{"Invalid category", "entertainment", CategoryNotAllowedError},
		{"Valid category", "food", nil},
	}

	for _, testCase := range testCases {
		t.Run(testCase.Name, func(t *testing.T) {
			t.Parallel()
			err := validateCategory(testCase.Category, allowed)
			require.ErrorIs(t, err, testCase.err)
		})
	}
}

func TestValidateTransaction(t *testing.T) {
	allowedTypes := []string{"income", "expense"}
	categories := []string{"food", "salary"}

	testCases := []struct {
		Name   string
		Body   web_helpers.TransactionRequest
		errLen int
	}{
		{
			"All correct",
			web_helpers.TransactionRequest{
				AccountId:   1, // ИСПРАВЛЕНО: добавили AccountId
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
				AccountId:   1, // ИСПРАВЛЕНО
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
				AccountId:   1, // ИСПРАВЛЕНО
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
				AccountId:   1, // ИСПРАВЛЕНО
				Title:       "Test",
				Description: "Desc",
				Value:       100,
				Type:        "wrong",
				Category:    "wrong",
			},
			2,
		},
		{
			"Invalid AccountId", // ИСПРАВЛЕНО: Заменили проверку валюты на проверку AccountId
			web_helpers.TransactionRequest{
				AccountId:   0, // Ошибка здесь
				Title:       "Test",
				Description: "Desc",
				Value:       100,
				Type:        "income",
				Category:    "salary",
			},
			1,
		},
		{
			"All wrong",
			web_helpers.TransactionRequest{
				AccountId:   0,       // 1 ошибка
				Title:       "",      // 2 ошибка
				Description: "",      // 3 ошибка
				Value:       -1,      // 4 ошибка
				Type:        "wrong", // 5 ошибка
				Category:    "wrong", // 6 ошибка
			},
			6,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.Name, func(t *testing.T) {
			t.Parallel()
			errs := ValidateTransaction(testCase.Body, allowedTypes, categories)
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
		errs := ValidateSignUpUser(body)
		require.Empty(t, errs)
	})

	t.Run("Multiple errors in sign up", func(t *testing.T) {
		body := web_helpers.SignupBodyRequest{
			Username:        "ab",
			Password:        "admin",
			ConfirmPassword: "wrong",
			Email:           "invalid",
		}
		errs := ValidateSignUpUser(body)
		require.Len(t, errs, 5)
	})
}

func TestValidateUpdateUser(t *testing.T) {
	t.Parallel()

	ptr := func(s string) *string { return &s }

	t.Run("No fields to update", func(t *testing.T) {
		body := web_helpers.UpdateUserProfileRequest{}
		errs := ValidateUpdateUser(body)
		require.Len(t, errs, 1)
		require.Equal(t, "request", errs[0].Field)
	})

	t.Run("All fields valid", func(t *testing.T) {
		body := web_helpers.UpdateUserProfileRequest{
			Username: ptr("newadmin"),
			Email:    ptr("new@example.com"),
			Password: ptr("NewAdmin123"),
		}
		errs := ValidateUpdateUser(body)
		require.Empty(t, errs)
	})

	t.Run("All fields invalid", func(t *testing.T) {
		body := web_helpers.UpdateUserProfileRequest{
			Username: ptr("ab"),
			Email:    ptr("invalid-email"),
			Password: ptr("weak"),
		}
		errs := ValidateUpdateUser(body)
		require.Len(t, errs, 3)
	})
}

func TestValidateLength(t *testing.T) {
	testCases := []struct {
		name      string
		data      string
		minLength int
		maxLength int
		err       error
	}{
		{
			name:      "valid",
			data:      "123",
			minLength: 1,
			maxLength: 10,
			err:       nil,
		},
		{
			name:      "too short",
			data:      "123",
			minLength: 5,
			maxLength: 10,
			err:       MinLengthError,
		},
		{
			name:      "too long",
			data:      "123123123123123123",
			minLength: 3,
			maxLength: 10,
			err:       MaxLengthError,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()
			err := validateLength(testCase.data, testCase.minLength, testCase.maxLength)
			require.Equal(t, testCase.err, err)
		})
	}
}

func TestCheckEqual(t *testing.T) {
	testCases := []struct {
		name   string
		first  any
		second any
		err    error
	}{
		{
			name:   "equal int",
			first:  10,
			second: 10,
			err:    nil,
		},
		{
			name:   "not equal int",
			first:  10,
			second: 1,
			err:    NoEqualsError,
		},
		{
			name:   "equal string",
			first:  "test",
			second: "test",
			err:    nil,
		},
		{
			name:   "not equal string",
			first:  "test",
			second: "123",
			err:    NoEqualsError,
		},
		{
			name:   "equal float",
			first:  1.0,
			second: 1.0,
			err:    nil,
		},
		{
			name:   "not equal float",
			first:  1.0,
			second: 1.1,
			err:    NoEqualsError,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()
			err := checkEqual(testCase.first, testCase.second)
			require.Equal(t, testCase.err, err)
		})
	}
}
