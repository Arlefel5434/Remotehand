# Этап 2. Docker-окружение

## Результат этапа

Все команды разработки выполняются через Docker Compose; контейнер также
кросс-компилирует нативный Windows-agent.

## Архитектура

```text
Docker Compose (WSL)
├── dev: format, vet, test
├── server: HTTP/WebSocket в mock-режиме
└── build-windows: создаёт dist/remotehand.exe

Windows host
└── dist/remotehand.exe: SendInput и Core Audio
```

## Шаги

1. Добавить multi-stage `Dockerfile` на закреплённой версии Go.
2. Добавить `.dockerignore`, исключив `.git`, `bin`, `dist` и временные файлы.
3. Создать `compose.yaml` с сервисами `dev`, `server` и `build-windows`.
4. Подключать исходники в `dev`, но использовать именованные volumes для Go
   build/module cache.
5. В `server` открыть порт `18080` и включить mock input adapter.
6. В `build-windows` установить `GOOS=windows`, `GOARCH=amd64` и складывать
   результат в `dist/`.
7. Обернуть частые команды в `Makefile`: `make test`, `make check`, `make dev`,
   `make build-windows`.
8. Закрепить версии базовых образов; не запускать процесс от root там, где это
   не требуется.

## Проверка

```bash
docker compose config
docker compose build
docker compose run --rm dev go test ./...
docker compose run --rm dev go vet ./...
docker compose run --rm build-windows
test -f dist/remotehand.exe
```

Затем запустить mock-сервер:

```bash
docker compose up --build server
```

`http://localhost:18080/health` должен отвечать, а `docker compose down` —
останавливать окружение без оставшихся контейнеров.

## Критерий готовности

Разработчику нужны только Git, Docker Desktop и Compose. Тесты и Windows-сборка
не зависят от Go, установленного в WSL.

## Коммит

`build: add Docker development environment`

