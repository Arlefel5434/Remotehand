# Этап 1. Репозиторий и окружение

## Результат этапа

Минимальный Go-проект собирается в WSL для Windows, тестируется и хранится в Git.

## Шаги

1. Проверить `go version` и `git --version`.
2. Инициализировать модуль: `go mod init github.com/<owner>/remotehand`.
3. Создать `cmd/remotehand/main.go` с минимальной функцией `main`.
4. Создать каталоги `internal/input`, `internal/protocol`, `internal/server`,
   `web`, `scripts` и `bin`.
5. Добавить `Makefile` или скрипты для тестов и Windows-сборки.
6. Выполнить `git init`, создать ветку `main` и первый коммит.
7. После получения URL добавить `origin` и выполнить push.

## Проверка

```bash
go test ./...
go build ./cmd/remotehand
GOOS=windows GOARCH=amd64 go build -o bin/remotehand.exe ./cmd/remotehand
git status --short
```

Ожидается: тесты и обе сборки успешны, `remotehand.exe` существует, рабочее
дерево после коммита чистое.

## Критерий готовности

Проект воспроизводимо собирается, а ветка `main` доступна в удалённом
репозитории.

## Коммит

`chore: initialize Go project`

