.DEFAULT: app
# phony specify that these keywords are commands
.PHONY: app api clean nuke gen plint

# Run app service (dependecies in docker-compose runs all related services)
app:
	@docker compose run --build --service-ports --rm app

# hot reload api build
api:
	@go run github.com/cespare/reflex@latest -d none -r 'api/.*' -s -- docker compose run --build --service-ports --rm api

# stop all containers
clean:
	@docker compose down

# Remove everything (good for fresh database / caching issues)
nuke:
	@docker composed down -v
	@docker system prune -a

# Generate proto code
gen:
	@buf generate 

# format proto files
plint:
	@clang-format -i proto/api/v1/*.proto
