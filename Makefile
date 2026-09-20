BIN := bin/ynafb-server
WEB := cmd/ynafb-server/web

.PHONY: build build-backend build-frontend run backend frontend clean gen gen-backend gen-frontend

## Production build: frontend bundle (embedded into the server) + Go server binary.
build: build-frontend build-backend

build-backend: gen-backend
	go build -o $(BIN) ./cmd/ynafb-server

build-frontend: gen-frontend
	npm --prefix frontend run build

## Run the dev environment: Go API server (:8080) + Vite dev server (:5173, proxies to :8080).
run:
	$(MAKE) -j2 backend frontend

backend: gen-backend ensure-web
	go run ./cmd/ynafb-server

# The server embeds cmd/ynafb-server/web; build it once if missing.
ensure-web:
	@test -d $(WEB) || $(MAKE) build-frontend

frontend: gen-frontend
	npm --prefix frontend run dev

## Regenerate all generated code: Go (oapi-codegen + sqlc) + frontend (openapi-typescript).
gen: gen-backend gen-frontend

## Go codegen: internal/http/api/server.gen.go from api.yml, data/ from queries/ + migrations/.
gen-backend:
	go generate ./...

## Frontend types: frontend/src/lib/api/schema.d.ts from api.yml.
gen-frontend:
	npm --prefix frontend run gen:api

clean:
	rm -rf $(BIN) $(WEB) frontend/dist
