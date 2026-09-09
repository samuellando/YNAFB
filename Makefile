BIN := bin/ynafb-server
WEB := cmd/ynafb-server/web

.PHONY: build build-backend build-frontend run backend frontend clean

## Production build: frontend bundle (embedded into the server) + Go server binary.
build: build-frontend build-backend

build-backend:
	go build -o $(BIN) ./cmd/ynafb-server

build-frontend:
	npm --prefix frontend run build

## Run the dev environment: Go API server (:8080) + Vite dev server (:5173, proxies to :8080).
run:
	$(MAKE) -j2 backend frontend

backend: ensure-web
	go run ./cmd/ynafb-server

# The server embeds cmd/ynafb-server/web; build it once if missing.
ensure-web:
	@test -d $(WEB) || $(MAKE) build-frontend

frontend:
	npm --prefix frontend run dev

clean:
	rm -rf $(BIN) $(WEB) frontend/dist
