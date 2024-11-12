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