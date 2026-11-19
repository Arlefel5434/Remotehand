.PHONY: dev fmt check test vet server down build-windows clean

dev:
	docker compose run --rm dev sh

fmt:
	docker compose run --rm dev gofmt -w .

check: fmt vet test build-windows

test:
	docker compose run --rm dev go test ./...

vet:
	docker compose run --rm dev go vet ./...

server:
	docker compose up --build server

down:
	docker compose down --remove-orphans

build-windows:
	docker compose run --rm --build build-windows

clean:
	rm -rf dist
