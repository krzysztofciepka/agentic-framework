.PHONY: build run test dev frontend clean

build: frontend
	go build -o bin/server ./cmd/server

run: build
	./bin/server

dev:
	go run ./cmd/server

frontend:
	cd web && npm install --silent && npm run build

test:
	go test ./...

clean:
	rm -rf bin/ web/dist/ web/node_modules/

.DEFAULT_GOAL := build
