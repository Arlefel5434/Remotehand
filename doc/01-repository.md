# Этап 1. Репозиторий и окружение

## Результат этапа

Минимальный Go-проект хранится в Git и подготовлен к контейнерной разработке.

## Шаги

1. Проверить `docker version`, `docker compose version` и `git --version`.
2. Инициализировать модуль: `go mod init github.com/<owner>/remotehand`.
3. Создать `cmd/remotehand/main.go` с минимальной функцией `main`.
4. Создать каталоги `internal/input`, `internal/protocol`, `internal/server`,
   `web`, `scripts` и `bin`.
5. Добавить `Makefile`, команды которого вызывают Docker Compose.
6. Выполнить `git init`, создать ветку `main` и первый коммит.
7. После получения URL добавить `origin` и выполнить push.

## Проверка

```bash
docker version
docker compose version
git status --short
```

Ожидается: Docker доступен из WSL, Compose установлен, рабочее дерево после
коммита чистое. Сборка проверяется на следующем этапе.

## Критерий готовности

Структура проекта создана, Docker доступен, ветка `main` опубликована.

## Коммит

`chore: initialize Go project`

