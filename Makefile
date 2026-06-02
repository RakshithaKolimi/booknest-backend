-include .env
export

# Migration path
MIGRATIONS_DIR=internal/http/database/migrations
MIGRATE=migrate -path $(MIGRATIONS_DIR) -database $(DB_URL)


# Commands
.PHONY: run
run:
	go run .

migrate-up:
	echo "Running migrations..."
	$(MIGRATE) up

migrate-down:
	echo "Rolling back one migration..."
	$(MIGRATE) down 1

migrate-down-all:
	echo "Rolling back all migrations..."
	$(MIGRATE) down
	
migrate-force:
	echo "Forcing version..."
	$(MIGRATE) force $(VERSION)

migrate-version:
	echo "Fetching version..."
	$(MIGRATE) version

migrate-new:
	migrate create -ext sql -dir $(MIGRATIONS_DIR)  $(name)

doc:
	swag init --parseDependency --parseInternal
