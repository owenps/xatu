.PHONY: run build lint test clean

run:
	go run ./cmd/xatu

build:
	go build -o bin/xatu ./cmd/xatu

lint:
	golangci-lint run ./...

test:
	go test ./...

clean:
	rm -rf bin/
