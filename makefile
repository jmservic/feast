DB_PROTOCOL := postgres 
DB_URL := postgres://postgres:postgres@localhost:5432/feast
GOOSE_CMD := goose $(DB_PROTOCOL) $(DB_URL)
SERVER_EXE := feast_server

define get-server-pid
ps | awk '/$(SERVER_EXE)/ {print $$1}'
endef

SERVER_PID := $(shell $(get-server-pid))

.PHONY: up down start stop test build unit

test:
	echo $(SERVER_PID) 

up:
	cd ./sql/schema; \
	$(GOOSE_CMD) up

down: 
	cd ./sql/schema; \
	$(GOOSE_CMD) down;

start: build
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

# Add file dependencies and take this out of PHONY
build:
	go build -o $(SERVER_EXE) ./cmd/server

#Fix this for brace expansion...
unit:
	go test -cover ./cmd/... ./internal/...

integration: start
	while [ -z "$$($(get-server-pid))" ]; do \
		echo "Waiting for Server to start..."; \
		sleep 1; \
	done;
	go test ./integration_tests/...
