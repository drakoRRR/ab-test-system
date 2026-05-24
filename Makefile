# Makefile (repo root)
# Central entry point for demo-related commands.

FIREBASE_TOKEN ?=
API_URL        ?= http://localhost:8080/api/v1
DEMO_DIR       ?= demo

.PHONY: demo-setup demo-run demo-k6

## demo-setup: create org/project/experiment/SDK key via admin API, write demo/.env
## Prerequisite: backend must be running and FIREBASE_TOKEN must be set.
##   export FIREBASE_TOKEN=<value from DevTools → Cookies → firebase-token>
demo-setup:
	@test -n "$(FIREBASE_TOKEN)" || \
	  (echo "ERROR: FIREBASE_TOKEN is not set."; \
	   echo "  1. Start the backend:  cd backend && make dev"; \
	   echo "  2. Log in to the admin UI at http://localhost:3000"; \
	   echo "  3. DevTools → Application → Cookies → copy 'firebase-token'"; \
	   echo "  4. export FIREBASE_TOKEN=<value>"; \
	   exit 1)
	cd demo/backend && FIREBASE_TOKEN=$(FIREBASE_TOKEN) API_URL=$(API_URL) DEMO_DIR=.. \
	  go run ./cmd/seed

## demo-run: start the demo HTTP server (requires demo/.env)
demo-run:
	@test -f demo/.env || (echo "ERROR: demo/.env not found. Run 'make demo-setup' first." && exit 1)
	cd demo/backend && set -a && . ../.env && set +a && go run .

## demo-k6: run the k6 traffic generator against the demo server
demo-k6:
	k6 run k6/scenarios/ab_test.js
