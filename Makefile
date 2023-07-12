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

.PHONY: deploy
deploy:
	go run cmd/deploy/deploy.go
