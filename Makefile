.PHONY: up down test console console-write up-prod console-prod console-prod-write console-symbols

# Start the whole stack: postgres, redis, migrations, seed, api with hot reload
up:
	docker-compose up --build -d

down:
	docker-compose down

# Tests: a one-off container from the api service definition; compose starts
# the database dependencies automatically
test:
	docker-compose run --rm api go test ./...

# Console (rails c analogue): a one-off container with the same environment;
# deps start automatically, the running api container is not touched.
# Read-only by default, -write to allow writes.
console:
	docker-compose run --rm api go run ./cmd/console

console-write:
	docker-compose run --rm api go run ./cmd/console -write

# Production-shaped stack: one image with both /app/api and /app/console,
# no source mounts, no hot reload (own project name, isolated from dev).
# Console access matches production: exec into the RUNNING api container —
# the deployment itself, not a side container.
up-prod:
	docker-compose -f docker-compose.prod.yml up --build -d

console-prod:
	docker-compose -f docker-compose.prod.yml exec api /app/console

console-prod-write:
	docker-compose -f docker-compose.prod.yml exec api /app/console -write

# Regenerate console symbol tables: make console-symbols DOMAIN=orders
# DOMAIN accepts one domain or a comma-separated list (DOMAIN=orders,catalog);
# omit it to regenerate every domain. Both packages of a domain are extracted:
# the domain itself (params, filters, errors) and its models (a model may be
# a method argument, so it must be constructible in the session). Extract
# writes one file per package straight into internal/console/symbols/; mv
# turns the raw names into <domain>.go and <domain>_models.go, overwriting
# the previous version. Raw extract keys are kept as is — session names are
# derived in symbols.Exports().
DOMAINS ?= orders catalog

console-symbols:
	@for domain in $$(echo $(if $(DOMAIN),$(DOMAIN),$(DOMAINS)) | tr ',' ' '); do \
		echo "==> $$domain: extracting domain + models"; \
		( cd internal/console/symbols && \
			go run github.com/traefik/yaegi/cmd/yaegi extract -name symbols \
				cosy-console/internal/domain/$$domain cosy-console/internal/domain/$$domain/models && \
			mv -f cosy-console-internal-domain-$$domain.go $$domain.go && \
			mv -f cosy-console-internal-domain-$$domain-models.go $${domain}_models.go ); \
		echo "==> $$domain: written $$domain.go + $${domain}_models.go"; \
	done
