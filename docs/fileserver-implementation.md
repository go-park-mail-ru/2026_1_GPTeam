# Fileserver: реализация и эксплуатация

Документ фиксирует итог выделения микросервиса раздачи и приёма файлов (аватары) из монолита. В коде реализации намеренно нет поясняющих комментариев к логике; детали сосредоточены здесь.

## Архитектура

- **Бинарь `cmd/fileserver`**: только HTTP.
  - `GET /healthz` — проверка готовности (ответ `204 No Content`).
  - `GET /img/{имяФайла}` — отдача из каталога хранилища с префиксом `/img/`, без листинга каталога при запросе ровно `/img/`.
  - `POST /internal/upload` — приём multipart (поле файла `file`, поле `extension` например `.png`). Требуется заголовок `Authorization: Bearer <FILESERVER_UPLOAD_TOKEN>`. Ответ JSON: `{"filename":"<uuid.ext>"}`.
- Если `FILESERVER_UPLOAD_TOKEN` не задан на стороне fileserver, `POST /internal/upload` отдаёт `503`; публичная отдача `/img/` продолжает работать при наличии файлов на диске.

**Бинарь BFF (`main.go`)**:

- Если задан **`FILESERVER_INTERNAL_URL`** (пробельные символы обрезаются), загрузка идёт HTTP-клиентом в fileserver; обязателен **`FILESERVER_UPLOAD_TOKEN`** (иначе при старте BFF — `Fatal`).
- Если **`FILESERVER_INTERNAL_URL` пуст**, используется **`internal/storage.LocalAvatar`**, пишущий в **`FILESERVER_STORAGE_PATH`** (по умолчанию `./static`), и BFF сам вешает маршрут **`/img/`** как раньше (локальная разработка без второго процесса).

**Слой application (`internal/application/user.go`)**:

- Зависимость **`AvatarUploader`** (`internal/application/avatar_uploader.go`): единая точка «сохранить байты, вернуть имя файла».
- После успешного `Upload` вызывается **`UpdateAvatar`** в PostgreSQL; в БД по-прежнему хранится только имя файла (ключ объекта).

**Клиент удалённой загрузки**: `internal/clients/fileserver/upload.go`, пакет `fileserver` (импорт в монолите с алиасом `fileupload`).

## URL аватара в ответе API

`internal/web/user_handler.go` после загрузки строит публичный URL:

1. Если задан **`FILESERVER_PUBLIC_BASE`** (без обязательного завершающего `/`) — база берётся оттуда.
2. Иначе — из **`SERVER_URL`** (поведение как раньше).
3. Итог: `JoinPath(publicBase, "img", имяФайла)`.

При docker-compose проброшен fileserver как **`8083:8082`**, по умолчанию для `app` задано **`FILESERVER_PUBLIC_BASE=${FILESERVER_PUBLIC_BASE:-http://localhost:8083}`**, чтобы браузер запрашивал картинки с нужного хоста и порта. Если перед внешним reverse proxy всё висит на одном origin, выставите **`FILESERVER_PUBLIC_BASE`** равным базовому URL того же приложения.

## Content-Security-Policy

`internal/secure/csp.go`: если задан **`FILESERVER_PUBLIC_BASE`**, из него извлекается origin (`scheme` + `host`) и добавляется к директиве **`img-src`** (наряду с `'self'` и `data:`). Иначе на чужой origin браузер бы заблокировал изображения.

## Переменные окружения

| Переменная | Где используется | Назначение |
|------------|-------------------|------------|
| `FILESERVER_LISTEN` | fileserver | Адрес прослушивания (`:8082` по умолчанию) |
| `FILESERVER_STORAGE_PATH` | fileserver, BFF (локальный режим) | Каталог файлов (`./static` по умолчанию) |
| `FILESERVER_READ_TIMEOUT_SEC` | fileserver | Таймаут чтения HTTP (секунды) |
| `FILESERVER_WRITE_TIMEOUT_SEC` | fileserver | Таймаут записи HTTP (секунды) |
| `FILESERVER_UPLOAD_TOKEN` | fileserver + BFF (удалённый режим) | Общий секрет Bearer для internal upload |
| `FILESERVER_INTERNAL_URL` | BFF | База URL без хвоста `/` (например `http://fileserver:8082`) для `POST …/internal/upload` |
| `FILESERVER_PUBLIC_BASE` | BFF | Базовый URL без хвоста `/` для ссылок и CSP на отображение аватаров |

Дополнительно по-прежнему используются **`SERVER_URL`**, если `FILESERVER_PUBLIC_BASE` не задан.

Образец значений см. [../.env.example](../.env.example).

## Docker и compose

[Dockerfile](../Dockerfile) собирает артефакты `app`, `auth-service`, `ai-service`, **`fileserver`**.

[docker-compose.yaml](../docker-compose.yaml):

- Сервис **`fileserver`**: `./fileserver`, именованный том **`fileserver_static:/app/static`**, внутренний порт `8082`, снаружи **`8083:8082`**.
- **`app`**: `FILESERVER_INTERNAL_URL=http://fileserver:8082`, `FILESERVER_PUBLIC_BASE` через подстановку; монтирование **`./static` с хоста снято** — файлы только в томе fileserver.
- Перед первым деплоем нужно добавить **`FILESERVER_UPLOAD_TOKEN`** в `.env` (одно и то же значение видят и `app`, и `fileserver` через `env_file`).

Зависимость **`depends_on fileserver`** у `app` гарантирует порядок старта; для готовности по health-check можно усилить позже.

## Middleware BFF

`AuthMiddleware` по-прежнему пропускает без JWT пути с префиксом **`/img/`** — безвредно, если монолит в локальном режиме снова отдаёт статику. В docker-режиме клиенты ходят на fileserver, а не на BFF для картинок.

## Лимиты и безопасность

- Размер тела internal upload ограничен константой **`MaxInternalUploadBody`** (`6 << 20` байт) в `internal/fileserver/http.go` и при копировании в файл.
- Проверка JWT на `GET /img/` не выполняется: доступ по угадыванию UUID имени (как и при старом `FileServer` на монолите). Ужесточение — отдельная задача (подписанные URL, отдельный приватный bucket).
- Сравнение токена через **`subtle.ConstantTimeCompare`**.

## Миграция данных и откат

- Существующие файлы из хостового `./static` при переходе на compose с томом не копируются автоматически: при необходимости залейте их в том `fileserver_static` или выполните разовый `docker cp`.
- Откат: убрать `FILESERVER_INTERNAL_URL` у BFF, вернуть volume `./static` на `app` и маршрут `/img/` снова появится в локальном режиме.

## Тесты

Юнит-тесты use case используют **`storage.NewLocalAvatar(t.TempDir())`** и не требуют запущенного fileserver. Для регрессии интеграции клиента можно поднять `httptest.Server` с `fileserver.NewRouter` в отдельном тесте (по желанию).

## Что намеренно не входит в scope

- Голосовые транзакции и Groq не используют fileserver (аудио обрабатывается в памяти).
- Нет gRPC-протокола fileserver: выбран HTTP multipart для простоты и отладки.
