build:
	@go build -o bin/blocker

run: build
	@./bin/blocker

test:
	@go test -v ./...

proto:
	##export PATH="$PATH:$(go env GOPATH)/bin"
	@protoc --go_out=. --go-grpc_out=. proto/*.proto


.PHONY: proto