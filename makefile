SHELL := /bin/bash
DB_PROTOCOL := postgres 
DB_URL := postgres://postgres:postgres@localhost:5432/feast
GOOSE_CMD := goose $(DB_PROTOCOL) $(DB_URL)
SERVER_EXE := feast_server
SRC_FOLDERS := cmd internal
SRC_FOLDERS_PATHS := $(addprefix ./,$(SRC_FOLDERS))
UNIT_TEST_ARGS := $(addsuffix /...,$(SRC_FOLDERS_PATHS))
SRC_FILES := $(foreach folder,$(SRC_FOLDERS_PATHS), \
				 $(shell find $(folder) -name *.go))

define get-server-pid
	ps | awk '/$(SERVER_EXE)/ {print $$1}'
endef

SERVER_PID := $(shell $(get-server-pid))

.PHONY: up down start stop test unit

$(SERVER_EXE): $(SRC_FILES)
	go build -o $(SERVER_EXE) ./cmd/server

up:
	cd ./sql/schema; \
	$(GOOSE_CMD) up

down: 
	cd ./sql/schema; \
	$(GOOSE_CMD) down;

start: $(SERVER_EXE)
	 @ if [ -n "$(SERVER_PID)" ]; then \
		echo "server is running"; \
	else \
		./$(SERVER_EXE) & \
	fi

stop:
	@ if [ -n "$(SERVER_PID)" ]; then \
		echo "killing the feast server"; \
		kill $(SERVER_PID); \
	fi

unit:
	go test -cover $(UNIT_TEST_ARGS) 

integration: start
	@ while [ -z "$$($(get-server-pid))" ]; do \
		echo "Waiting for Server to start..."; \
		sleep 1; \
	done;
	go test ./integration_tests/...
ifndef SERVER_PID
	kill $(shell $(get-server-pid))
endif

test:
	@echo $(SERVER_PID) 
	@echo $(SRC_FILES)
