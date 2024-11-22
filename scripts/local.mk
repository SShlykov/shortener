ifneq (,$(wildcard example.env))
    include example.env
    export $(shell sed 's/=.*//' example.env)
endif

.PHONY: run
run:
	go run cmd/shortener/main.go

.PHONY: lint
lint:
	golangci-lint run

.PHONY: migrate
migrate:
	docker-compose up -d --force-recreate --build --remove-orphans && \
	go run cmd/migrator/main.go up

.PHONY: .sqlc
.sqlc:
	sqlc generate -f ./sqlc/sqlc.json