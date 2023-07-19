.PHONY: generate
generate:
	go generate ./...

.PHONY: install
install: generate
install:
	go mod download
	go install ./...

.PHONY: format
format:
	go fmt ./...

.PHONY: build
build:
	go run cmd/build/build.go ${DUMP_PATH} docs

.PHONY: deploy
deploy:
	go run cmd/deploy/deploy.go
