DB_PROTOCOL := postgres 
DB_URL := postgres://postgres:postgres@localhost:5432/feast
GOOSE_CMD := goose $(DB_PROTOCOL) $(DB_URL)

.PHONY: up down

up:
	cd ./sql/schema; \
	$(GOOSE_CMD) up

down: 
	cd ./sql/schema; \
	$(GOOSE_CMD) down;
