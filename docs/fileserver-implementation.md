# Fileserver

Микросервис, который принимает аватары пользователей и отдаёт их по HTTP. Основной сервис общается с ним по gRPC (по аналогии с `auth` и `ai`); HTTP у fileserver остаётся только для отдачи `/img/` и health-check.

## Компоненты

`cmd/fileserver` запускает два листенера:

- **gRPC** (`FILESERVER_GRPC_LISTEN`, по умолчанию `:50053`) — `FileService.Upload(bytes, extension) -> filename`.
- **HTTP** (`FILESERVER_HTTP_LISTEN`, по умолчанию `:8082`) — `GET /img/{name}` и `GET /healthz`.

Внутри `internal/fileserver` три слоя:

```
grpcserver  — gRPC-хендлер, валидирует запрос, мапит ошибки на gRPC-коды
application — use case AvatarService: лимит размера, нормализация расширения
storage     — LocalStorage: запись в каталог на диске под уникальным именем
```

`grpcserver` ничего не знает про файлы, `storage` — про gRPC.

## Контракт

`proto/fileserver/v1/fileserver.proto`:

```
service FileService {
  rpc Upload(UploadRequest) returns (UploadResponse);
}
message UploadRequest  { bytes data = 1; string extension = 2; }
message UploadResponse { string filename = 1; }
```

Лимит тела — `MaxUploadBytes = 6<<20` (на стороне сервиса). Ошибки `ErrEmptyData`/`ErrTooLarge` мапятся в `codes.InvalidArgument`, всё остальное — в `codes.Internal`.

## Клиент в основном сервисе

`internal/clients/fileserver.GrpcUploader` реализует интерфейс `application.AvatarUploader`. Основной сервис принимает multipart на `/api/profile/avatar`, читает байты файла и отправляет их одним gRPC-вызовом — без повторной упаковки в multipart.

Адрес fileserver задаётся переменной `FILESERVER_GRPC_ADDR` (например `fileserver:50053`).

## Публичные ссылки на аватары

`internal/web/user_handler.go` после успешной загрузки строит URL картинки:

1. база — `FILESERVER_PUBLIC_BASE`, если задан, иначе `SERVER_URL`;
2. итог — `JoinPath(base, "img", filename)`.

Если база отличается от origin основного сервиса, её origin добавляется в `img-src` CSP (`internal/secure/csp.go`).

## Переменные окружения

| Переменная | Где | Назначение |
|---|---|---|
| `FILESERVER_GRPC_LISTEN` | fileserver | адрес gRPC (`:50053`) |
| `FILESERVER_HTTP_LISTEN` | fileserver | адрес HTTP `/img`/`/healthz` (`:8082`) |
| `FILESERVER_STORAGE_PATH` | fileserver | каталог файлов (`./static`) |
| `FILESERVER_READ_TIMEOUT_SEC` / `FILESERVER_WRITE_TIMEOUT_SEC` | fileserver | таймауты HTTP |
| `FILESERVER_GRPC_ADDR` | основной сервис | адрес fileserver gRPC |
| `FILESERVER_PUBLIC_BASE` | основной сервис | база для публичного URL картинок и CSP |

## Docker

Все сервисы в `docker-compose.yaml` имеют `healthcheck` (`curl` для HTTP, `nc -z` для gRPC); основной сервис ждёт `service_healthy` от `auth`, `ai` и `fileserver`. Том `fileserver_static` хранит файлы только внутри fileserver, том на `app` снят.

## Тесты

- `internal/fileserver/storage` — IO в `t.TempDir()`, отдельный кейс на отмену контекста и ошибку чтения (партиальный файл удаляется).
- `internal/fileserver/application` — нормализация расширения, лимиты, прокидывание ошибок storage.
- `internal/fileserver/grpcserver` — мапинг ошибок use case на `codes.InvalidArgument`/`codes.Internal`.
- `internal/fileserver/httpserver` — `/healthz`, запрет листинга `/img/`, отдача файла из тестового каталога.
- `internal/clients/fileserver` — gRPC-клиент против поднятого in-process gRPC-сервера (`net.Listen` + `grpc.NewServer`).
- `internal/application` — `User.UploadAvatar` использует мок `AvatarUploader` (`go.uber.org/mock`), отдельная реализация на диске для прода больше не нужна.
