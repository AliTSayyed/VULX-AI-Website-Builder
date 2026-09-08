.DEFAULT: app
# phony specify that these keywords are commands
.PHONY: app api clean nuke gen plint trust trust-rm

# Run app service (dependecies in docker-compose runs all related services)
app: clean
	@docker compose run --build --service-ports --use-aliases --rm app

# hot reload api build
api: clean
	@go run github.com/cespare/reflex@latest -d none -r 'api/.*' -s -- docker compose run --build --service-ports --use-aliases --rm api

# stop all containers
clean:
	@docker compose down

# Remove everything (good for fresh database / caching issues)
nuke:
	@docker compose down -v
	@docker system prune -a

# Generate proto code
gen:
	@buf generate 

# format proto files
plint:
	@clang-format -i proto/api/v1/*.proto

# trust caddy's local CA in the macOS system keychain (macOS only, requires stack running)
trust:
	@docker compose cp caddy:/data/caddy/pki/authorities/local/root.crt ./caddy-root.crt
	@sudo security add-trusted-cert -d -r trustRoot -k /Library/Keychains/System.keychain ./caddy-root.crt
	@rm ./caddy-root.crt

# remove caddy's local CA from the macOS system keychain
trust-rm:
	@sudo security delete-certificate -c "Caddy Local Authority - 2026 ECC Root" /Library/Keychains/System.keychain
