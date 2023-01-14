BUILD_VERSION := $(if ${BUILD_VERSION},${BUILD_VERSION},$(shell git rev-parse --abbrev-ref HEAD))
BUILD_DATE := $(shell date +%FT%T%z)
LDFLAGS := -X github.com/bohdanch-w/go-tgupload/internal/build.Version=${BUILD_VERSION} -X github.com/bohdanch-w/go-tgupload/internal/build.Date=${BUILD_DATE}
build:
	@go build -ldflags "$(LDFLAGS)" -o bin/tg-upload.exe cmd/main.go

lint:
	@bin\golangci-lint run -c .golangci.yml ./...
