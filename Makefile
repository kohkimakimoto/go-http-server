.PHONY: default build fmt test testv deps deps_update

default: build

build:
	GOOS=darwin GOARCH=amd64 go build -o go-http-server-dawin
	GOOS=linux GOARCH=amd64 go build -o go-http-server-linux

fmt:
	go fmt $$(go list ./... | grep -v vendor)

test:
	go test -cover $(go list ./... | grep -v vendor)

testv:
	go test -cover -v $(go list ./... | grep -v vendor)

deps:
	gom install

deps_update:
	rm Gomfile.lock; rm -rf vendor; gom install && gom lock

