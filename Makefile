all: clean build test

cibuild: all

deps:
	go get github.com/zorkian/go-datadog-api

test-deps:
	go get -d github.com/stretchr/testify/assert \
	  golang.org/x/tools/cmd/cover

clean:
	rm -f ./datadog

./datadog: deps
	go build

build: ./datadog

fmt:
	go fmt

vet:
	go vet

smoke-test: ./datadog
	./datadog -dry-run gauge vsco.my_metric 100
	./datadog -dry-run incr vsco.my_metric 100

test: test-deps fmt vet smoke-test
	go test -cover
