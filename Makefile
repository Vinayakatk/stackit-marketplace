.PHONY: fmt run db redis start stop clean

# Format the Go code
fmt:
	go fmt ./...

# Function to check and restart a container
define restart_container
	@if [ -n "$$(docker ps -aq -f name=$(1))" ]; then \
		if [ -n "$$(docker ps -aq -f name=$(1) -f status=exited)" ]; then \
			echo "Removing exited container: $(1)"; \
			docker rm $(1); \
		else \
			echo "Container $(1) already running."; \
			exit 0; \
		fi \
	fi; \
	echo "Starting $(1) container..."; \
	$(2)
endef

# Check and start the database container
db:
	$(call restart_container,marketplace-db,\
		docker run --name marketplace-db \
		-e POSTGRES_USER=admin \
		-e POSTGRES_PASSWORD=secret \
		-e POSTGRES_DB=marketplace \
		-p 5433:5432 -d postgres)

# Check and start the Redis container
redis:
	$(call restart_container,marketplace-redis,\
		docker run --name marketplace-redis -p 6370:6379 -d redis)

# Set up environment, start DB/Redis if needed, and run the application
run: db redis
	@export DATABASE_URL="postgres://admin:secret@localhost:5433/marketplace"; \
	go run main.go

# Stop and remove the containers
stop:
	@docker stop marketplace-db marketplace-redis || true
	@docker rm marketplace-db marketplace-redis || true

# Clean up stopped containers
clean: stop
	@docker system prune -f
