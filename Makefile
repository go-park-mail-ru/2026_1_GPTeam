.PHONY: test
test:
	go test $$(go list ./... | grep -v '/mocks' | grep -v '/db' | grep -v '/models' | grep -v '/web_helpers') -v

.PHONY: test-cover
test-cover:
	go test $$(go list ./... | grep -v '/mocks' | grep -v '/db' | grep -v '/models' | grep -v '/web_helpers') -coverprofile=coverage.out -covermode=atomic

.PHONY: test-cover-html
test-cover-html: test-cover
	go tool cover -html=coverage.out -o coverage.html
	@xdg-open coverage.html 2>/dev/null || open coverage.html 2>/dev/null || echo "откройте coverage.html вручную"
	@go tool cover -func=coverage.out | grep total

.PHONY: clean-test
clean-test:
	rm -f coverage.out coverage.html