SHELL = /bin/sh -e
TAGS ?=
BIN_DIR ?= $(GOPATH)/bin

all: check_go_fmt deps test build

deps:
	if [ $$(uname) = "Linux" ]; then \
		wget -qO $(BIN_DIR)/dep https://github.com/golang/dep/releases/download/v0.3.2/dep-linux-amd64; \
		chmod +x $(BIN_DIR)/dep; \
	elif [ $$(uname) = "Darwin" ]; then \
		curl -sLo $(BIN_DIR)/dep https://github.com/golang/dep/releases/download/v0.3.2/dep-darwin-amd64; \
		chmod +x $(BIN_DIR)/dep; \
	else \
		>&2 echo "Unsupported OS: $$(uname)"; \
		exit 1; \
	fi
	dep ensure -vendor-only

test:
	go vet -tags=$(TAGS) ./...
	go test -tags=$(TAGS) ./...

build: clean
	mkdir build

	CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -a \
		-ldflags "-extldflags \"-static\" -X main.version=$${VERSION:-$${TRAVIS_TAG-}} -X main.gitHash=$${GIT_HASH:-$${TRAVIS_COMMIT-}} -X main.buildDate=$$(date -u +%F)" \
		-tags=$(TAGS) -o build/migrate github.com/pasztorpisti/migrate/cmd/migrate
	cd build \
		&& zip -q migrate-linux-amd64.zip migrate \
		&& shasum -a 256 migrate migrate-linux-amd64.zip > migrate-linux-amd64.zip.sha256 \
		&& rm migrate

	CGO_ENABLED=0 GOOS=darwin GOARCH=amd64 go build -a \
		-ldflags "-extldflags \"-static\" -X main.version=$${VERSION:-$${TRAVIS_TAG-}} -X main.gitHash=$${GIT_HASH:-$${TRAVIS_COMMIT-}} -X main.buildDate=$$(date -u +%F)" \
		-tags=$(TAGS) -o build/migrate github.com/pasztorpisti/migrate/cmd/migrate
	cd build \
		&& zip -q migrate-darwin-amd64.zip migrate \
		&& shasum -a 256 migrate migrate-darwin-amd64.zip > migrate-darwin-amd64.zip.sha256 \
		&& rm migrate

	CGO_ENABLED=0 GOOS=windows GOARCH=amd64 go build -a \
		-ldflags "-extldflags \"-static\" -X main.version=$${VERSION:-$${TRAVIS_TAG-}} -X main.gitHash=$${GIT_HASH:-$${TRAVIS_COMMIT-}} -X main.buildDate=$$(date -u +%F)" \
		-tags=$(TAGS) -o build/migrate.exe github.com/pasztorpisti/migrate/cmd/migrate
	cd build \
		&& zip -q migrate-windows-amd64.zip migrate.exe \
		&& shasum -a 256 migrate.exe migrate-windows-amd64.zip > migrate-windows-amd64.zip.sha256 \
		&& rm migrate.exe

clean:
	rm -rf build

check_go_fmt:
	@if [ -n "$$(gofmt -d $$(find . -name '*.go' -not -path './vendor/*'))" ]; then \
		>&2 echo "The .go sources aren't formatted. Please format them with 'go fmt'."; \
		exit 1; \
	fi

.PHONY: all deps test build clean check_go_fmt
