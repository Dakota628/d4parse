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

.PHONY: check-dump-path
check-dump-path:
	test -d ${DUMP_PATH} || { echo "'${DUMP_PATH}' does not exist!"; exit 1; }

.PHONY: build-docs
build: check-dump-path generate
	go run cmd/build/build.go ${DUMP_PATH} docs

.PHONY: build-map
build-map: check-dump-path
	go run cmd/dumpmap/dumpmap.go ${DUMP_PATH}
	go run cmd/dumpmapdata/dumpmapdata.go ${DUMP_PATH}
	cd map && npm run build

.PHONY: deploy
deploy:
	go run cmd/deploy/deploy.go

.PHONY: build-all
build-all: build-map build-docs
