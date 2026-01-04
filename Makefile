.PHONY: build run-server run-web clean

build:
	go build -o bin/server cmd/server/main.go

run-server:
	go run cmd/server/main.go

run-web:
	cd web && npm run dev

clean:
	rm -rf bin/
