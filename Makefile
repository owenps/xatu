.PHONY: run build lint test clean

run:
	go run .

build:
	go build -o bin/xatu .

lint:
	golangci-lint run ./...

test:
	go test ./...

clean:
	rm -rf bin/
