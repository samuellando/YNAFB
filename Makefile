BIN := bin/ynafb-server

.PHONY: build build-backend build-frontend run backend frontend clean

## Production build: Go server binary + frontend bundle.
build: build-backend build-frontend

build-backend:
	go build -o $(BIN) ./cmd/ynafb-server

build-frontend:
	npm --prefix frontend run build

## Run the dev environment: Go API server (:8080) + Vite dev server (:5173, proxies to :8080).
run:
	$(MAKE) -j2 backend frontend

backend:
	go run ./cmd/ynafb-server

frontend:
	npm --prefix frontend run dev

clean:
	rm -rf $(BIN) frontend/dist
