#!/bin/bash
set -e

# Сначала откатить файлы до чистого состояния
git checkout -- \
  ./internal/clients/groq/groq_client.go \
  ./internal/ai/groq/groq_client.go \
  ./internal/repository/rate_limiter.go \
  ./internal/repository/rate_limiter_test.go \
  ./internal/web/budget_handler_test.go \
  ./internal/web/transaction_handler_test.go \
  ./internal/web/user_handler_test.go \
  ./internal/clients/groq/groq_client_test.go \
  ./internal/web/support_handler_test.go \
  ./internal/web/auth_handler_test.go

EASYJSON="$(go env GOPATH)/bin/easyjson"

replace_json() {
  local f="$1"
  local keep_decoder="${2:-no}"  # yes = файл использует json.NewDecoder

  if [ "$keep_decoder" = "yes" ]; then
    # Добавить easyjson рядом с encoding/json (не удалять)
    grep -q 'mailru/easyjson' "$f" || \
      sed -i 's|"encoding/json"|"encoding/json"\n\teasyjson "github.com/mailru/easyjson"|' "$f"
  else
    # Полная замена импорта
    grep -q 'mailru/easyjson' "$f" || \
      sed -i 's|"encoding/json"|easyjson "github.com/mailru/easyjson"|' "$f"
  fi

  # \b — word boundary: не тронет easyjson.Marshal, только json.Marshal
  sed -i 's/\bjson\.Marshal(/easyjson.Marshal(/g'   "$f"
  sed -i 's/\bjson\.Unmarshal(/easyjson.Unmarshal(/g' "$f"
}

# Производственные файлы — полная замена
replace_json "./internal/clients/groq/groq_client.go"
replace_json "./internal/ai/groq/groq_client.go"
replace_json "./internal/repository/rate_limiter.go"

# Тестовые без NewDecoder — полная замена
replace_json "./internal/repository/rate_limiter_test.go"
replace_json "./internal/web/budget_handler_test.go"
replace_json "./internal/web/transaction_handler_test.go"

# user_handler_test.go: Unmarshal(&anonymous struct{}) оставить encoding/json
grep -q 'mailru/easyjson' "./internal/web/user_handler_test.go" || \
  sed -i 's|"encoding/json"|"encoding/json"\n\teasyjson "github.com/mailru/easyjson"|' "./internal/web/user_handler_test.go"
sed -i 's/\bjson\.Marshal(/easyjson.Marshal(/g' "./internal/web/user_handler_test.go"

# Тестовые с NewDecoder — два импорта
replace_json "./internal/clients/groq/groq_client_test.go"  "yes"
replace_json "./internal/web/support_handler_test.go"        "yes"
replace_json "./internal/web/auth_handler_test.go"           "yes"

# testhelper.go — НЕ ТРОГАЕМ: Marshal(value any) несовместим с easyjson.Marshaler

# Установка
go get github.com/mailru/easyjson
go install github.com/mailru/easyjson/easyjson@latest

# Кодогенерация
"$EASYJSON" -all ./internal/clients/groq/groq_client.go
"$EASYJSON" -all ./internal/ai/groq/groq_client.go
"$EASYJSON" -all ./internal/application/models/
"$EASYJSON" -all ./internal/web/web_helpers/
