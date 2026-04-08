EXCLUDE_DIRS := grep -v '/mocks' | grep -v '/db' | grep -v '/models' | grep -v '/web_helpers' | grep -vE '/pkg$$' | grep -vE 'github.com/go-park-mail-ru/2026_1_GPTeam$$'

.PHONY: mocks
mocks:
	@echo "Генерация моков для application..."
	mockgen -source=internal/application/account.go -destination=internal/application/mocks/mock_account.go -package=mocks
	mockgen -source=internal/application/user.go -destination=internal/application/mocks/mock_user.go -package=mocks
	mockgen -source=internal/application/budget.go -destination=internal/application/mocks/mock_budget.go -package=mocks
	mockgen -source=internal/application/transaction.go -destination=internal/application/mocks/mock_transaction.go -package=mocks
	mockgen -source=internal/application/enums.go -destination=internal/application/mocks/mock_enums.go -package=mocks
	mockgen -source=internal/application/voice_transaction.go -destination=internal/application/mocks/mock_voice_transaction.go -package=mocks
	
	@echo "Генерация моков для repository..."
	mockgen -source=internal/repository/account.go -destination=internal/repository/mocks/mock_account_repo.go -package=mocks
	mockgen -source=internal/repository/user.go -destination=internal/repository/mocks/mock_user_repo.go -package=mocks
	mockgen -source=internal/repository/budget.go -destination=internal/repository/mocks/mock_budget_repo.go -package=mocks
	mockgen -source=internal/repository/transaction.go -destination=internal/repository/mocks/mock_transaction_repo.go -package=mocks
	mockgen -source=internal/repository/enums.go -destination=internal/repository/mocks/mock_enums_repo.go -package=mocks
	mockgen -source=internal/repository/jwt.go -destination=internal/repository/mocks/mock_jwt_repo.go -package=mocks
	
	@echo "Генерация моков для auth..."
	mockgen -source=internal/auth/auth.go -destination=internal/auth/mocks/mock_auth.go -package=mocks
	mockgen -source=internal/auth/jwt_auth/jwt.go -destination=internal/auth/jwt_auth/mocks/mock_jwt.go -package=mocks
	
	@echo "Генерация моков для secure..."
	mockgen -source=internal/secure/csrf.go -destination=internal/secure/mocks/mock_csrf.go -package=mocks
	
	@echo "Готово!"

.PHONY: test
test:
	go test $$(go list ./... | $(EXCLUDE_DIRS)) -v

.PHONY: test-cover
test-cover:
	@echo "Запуск тестов с покрытием..."
	go test $$(go list ./... | $(EXCLUDE_DIRS)) -coverprofile=coverage.tmp -covermode=atomic
	@cat coverage.tmp | $(EXCLUDE_DIRS) > coverage.out
	@rm coverage.tmp

.PHONY: test-cover-html
test-cover-html: test-cover
	go tool cover -html=coverage.out -o coverage.html
	@xdg-open coverage.html 2>/dev/null || open coverage.html 2>/dev/null || echo "Откройте coverage.html вручную"
	@echo "------------------------------------------------------------------------"
	@go tool cover -func=coverage.out | grep total

.PHONY: clean-test
clean-test:
	rm -f coverage.out coverage.html coverage.tmp