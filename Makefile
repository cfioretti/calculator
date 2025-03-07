build:
	go build cmd/main.go

run:
	go run cmd/main.go

unit-test:
	go test -v ./internal/...

integration-test:
	go test -v ./test/...

test: unit-test integration-test
