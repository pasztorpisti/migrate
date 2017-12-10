SHELL = /bin/bash
TAGS ?=

all: deps test build

deps:
	if [[ $$(uname) == "Linux" ]]; then \
		wget -q -O $${GOPATH}/bin/dep https://github.com/golang/dep/releases/download/v0.3.2/dep-linux-amd64; \
		chmod +x $${GOPATH}/bin/dep; \
	elif [[ $$(uname) == "Darwin" ]]; then \
		curl -s -L -o $${GOPATH}/bin/dep https://github.com/golang/dep/releases/download/v0.3.2/dep-darwin-amd64; \
		chmod +x $${GOPATH}/bin/dep; \
	else \
		>&2 echo "Unsupported OS: $$(uname)"; \
		exit 1; \
	fi
	dep ensure -vendor-only

test:
	go test -tags=$(TAGS) ./...

build: clean
	mkdir build

	GOOS=linux GOARCH=amd64 go build -tags=$(TAGS) -o build/migrate github.com/pasztorpisti/migrate/cmd/migrate
	cd build \
	&& zip -q migrate-linux-amd64.zip migrate \
	&& shasum -a 256 migrate migrate-linux-amd64.zip > migrate-linux-amd64.zip.sha256 \
	&& rm migrate

	GOOS=darwin GOARCH=amd64 go build -tags=$(TAGS) -o build/migrate github.com/pasztorpisti/migrate/cmd/migrate
	cd build \
	&& zip -q migrate-darwin-amd64.zip migrate \
	&& shasum -a 256 migrate migrate-darwin-amd64.zip > migrate-darwin-amd64.zip.sha256 \
	&& rm migrate

	GOOS=windows GOARCH=amd64 go build -tags=$(TAGS) -o build/migrate.exe github.com/pasztorpisti/migrate/cmd/migrate
	cd build \
	&& zip -q migrate-windows-amd64.zip migrate.exe \
	&& shasum -a 256 migrate.exe migrate-windows-amd64.zip > migrate-windows-amd64.zip.sha256 \
	&& rm migrate.exe

clean:
	rm -rf build

.PHONY: all deps test build clean
