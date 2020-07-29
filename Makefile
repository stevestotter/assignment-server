export GO111MODULE=on

GOFILES= $$(go list -f '{{join .GoFiles " "}}')

.PHONY: mocks

export GO111MODULE=on

GOFILES= $$(go list -f '{{join .GoFiles " "}}')

.PHONY: mocks

deps:
	go mod vendor

mocks:
	rm -rf mocks
	go generate -v ./...

run:
	go run $(GOFILES)

test:
	go test -race -count=1 ./...