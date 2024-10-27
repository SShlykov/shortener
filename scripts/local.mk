ifneq (,$(wildcard example.env))
    include example.env
    export $(shell sed 's/=.*//' example.env)
endif

.PHONY: run
run:
	@echo "Hello World!"