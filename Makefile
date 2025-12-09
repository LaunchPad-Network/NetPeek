OUTPUT_DIR ?= bin
GO ?= go

GOOS ?= $(shell go env GOOS)
GOARCH ?= $(shell go env GOARCH)

ENTRYPOINTS = \
	bird-lg-go=cmd/bird-lg-go/main.go \
	bird-lgproxy-go=cmd/bird-lgproxy-go/main.go

TARGETS = \
	linux/amd64 \
	darwin/amd64 \
	linux/arm64 \
	darwin/arm64

EXT := $(if $(filter windows,$(GOOS)),.exe,)

NAME ?= bird-lg-go
SRC := $(word 2,$(subst =, ,$(filter $(NAME)=%,$(ENTRYPOINTS))))

BUILD_TIME := $(shell date +%Y-%m-%dT%H:%M:%S%z)
COMMIT_HASH := $(shell git rev-parse HEAD)

LDFLAGS := -ldflags "-X github.com/LaunchPad-Network/NetPeek/internal/version.entryPoint=$(NAME) -X github.com/LaunchPad-Network/NetPeek/internal/version.buildTime=$(BUILD_TIME) -X github.com/LaunchPad-Network/NetPeek/internal/version.commitHash=$(COMMIT_HASH)"

.PHONY: all run test build build-all cross cross-all clean codegen

all: build

run:
	$(GO) run $(LDFLAGS) $(SRC)

test:
	$(GO) test $(LDFLAGS) ./... -v

build:
	GOOS=$(GOOS) GOARCH=$(GOARCH) $(GO) build $(LDFLAGS) -o $(OUTPUT_DIR)/$(NAME)_$(GOOS)_$(GOARCH)$(EXT) $(SRC)

build-all:
	$(foreach pair, $(ENTRYPOINTS), \
		$(eval NAME := $(word 1,$(subst =, ,$(pair)))) \
		$(MAKE) build NAME=$(NAME); \
	)

cross:
	$(foreach target,$(TARGETS), \
		$(eval OS := $(word 1,$(subst /, ,$(target)))) \
		$(eval ARCH := $(word 2,$(subst /, ,$(target)))) \
		GOOS=$(OS) GOARCH=$(ARCH) $(MAKE) build NAME=$(NAME); \
	)

cross-all:
	$(foreach pair, $(ENTRYPOINTS), \
		$(eval NAME := $(word 1,$(subst =, ,$(pair)))) \
		$(foreach target,$(TARGETS), \
			$(eval OS := $(word 1,$(subst /, ,$(target)))) \
			$(eval ARCH := $(word 2,$(subst /, ,$(target)))) \
			GOOS=$(OS) GOARCH=$(ARCH) $(MAKE) build NAME=$(NAME); \
		) \
	)

clean:
	rm -rf $(OUTPUT_DIR)/*
