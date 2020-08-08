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

# Run unit tests (tests that aren't skipped on short flag)
test:
	go test -race -count=1 -short -timeout 10s ./...

# Run integration tests (tests with Integration in their name)
test-integration:
	go test -count=1 -run 'Integration' -timeout 30s ./...

docker-up:
	docker-compose up --build -d

docker-down:
	docker-compose down

docker-test: 
	docker-compose up --build -d 
	docker exec -it assignment-server sh -c '\
		make test && make test-integration'
	docker-compose down
