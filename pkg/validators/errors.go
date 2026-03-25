package validators

import "fmt"

var (
	ServerError = fmt.Errorf("ошибка сервера")

	UsernameShortError        = fmt.Errorf("логин должен быть минимум 3 символа")
	UsernameWrongSymbolsError = fmt.Errorf("логин должен содержать только буквы латинского алфавита или цифры")
	IncorrectPasswordError    = fmt.Errorf("пароль должен содержать заглавные, строчные буквы латинского алфавита и цифры (не менее 8 символов)")
	PasswordsNotSameError     = fmt.Errorf("пароли не совпадают")
	EmailError                = fmt.Errorf("некорректный адрес электронной почты")

	BudgetTitleEmpty        = fmt.Errorf("заголовок пустой")
	BudgetTitleTooLong      = fmt.Errorf("заголовок не может быть длиннее 255 символов")
	BudgetDescriptionEmpty  = fmt.Errorf("описание пустое")
	CurrencyNotAllowedError = fmt.Errorf("валюта не поддерживается")
	TargetIsNegativeError   = fmt.Errorf("планируемый бюджет не может быть меньше 0")
	StartDateInPastError    = fmt.Errorf("дата начала не может быть в прошлом")
	EndDateInPastError      = fmt.Errorf("дата окончания должна быть позже даты начала")

	ValueIsNegativeError = fmt.Errorf("значение не может быть меньше нуля")
	ValueIsBigError      = fmt.Errorf("значение не может быть больше 1e12")

	TransactionTypeNotAllowedError     = fmt.Errorf("тип транзакции не поддерживается")
	TransactionCategoryNotAllowedError = fmt.Errorf("категория транзакции не поддерживается")
	TransactionTitleEmptyError         = fmt.Errorf("заголовок пустой")
	TransactionTitleLongError          = fmt.Errorf("заголовок не может быть длиннее 255 символов")
	TransactionDescriptionEmptyError   = fmt.Errorf("описание пустое")
)
