ifeq ($(OS),Windows_NT)
	NOW := $(shell powershell -NoProfile -Command "Get-Date -Format 'yyyy-MM-dd_HH-mm-ss'")
	NULL := NUL
	OUT_FILE := gotg.exe
else
	NOW := $(shell date +'%Y-%m-%d_%H-%M-%S')
	NULL := /dev/null
	OUT_FILE := gotg
endif

VERSION := $(strip $(shell \
	git describe --tags --exact-match 2> $(NULL) || \
	( \
		BR=$$(git rev-parse --abbrev-ref HEAD 2> $(NULL)); \
		HASH=$$(git rev-parse --short HEAD 2> $(NULL)); \
		if [ "$$BR" = "HEAD" ]; then \
			echo "$$HASH"; \
		else \
			echo "$$BR-$$HASH"; \
		fi \
	) \
))

GO_VERSION := $(shell go version | awk '{print $$3}')

LDFLAGS :=  -X github.com/bohdanch-w/go-tgupload/internal/build.Version=$(VERSION) \
			-X github.com/bohdanch-w/go-tgupload/internal/build.GoVersion=$(GO_VERSION) \
			-X github.com/bohdanch-w/go-tgupload/internal/build.BuiltAt=$(NOW)

build:
	go build -ldflags "$(LDFLAGS)" -o bin/$(OUT_FILE) cmd/main.go

version:
	@echo Version: $(VERSION)
	@echo BuiltAt $(NOW)
	@echo GoVersion $(GO_VERSION)
	@echo Flags $(LDFLAGS)

lint:
	@bin\golangci-lint run -c .golangci.yml ./...
