# export PATH="$PATH:$(go env GOPATH)/bin"

build:
	@go build -o bin/blocker

run: build
	@./bin/blocker

test:
	@go test -v ./... -count=1

proto:
	@protoc --go_out=. --go-grpc_out=. proto/*.proto


.PHONY: proto
.PHONY: test