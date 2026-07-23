BINARY := md

.PHONY: all build test vet fmt clean release

# Default target: bare `make` builds the binary.
all: build

# Compile the md binary into the repo root.
build:
	go build -o $(BINARY) ./cmd/md

test:
	go test ./...

vet:
	go vet ./...

fmt:
	gofmt -w .

clean:
	rm -f $(BINARY)
	rm -rf dist

# Cut a release tag. VERSION takes patch, minor, major, or an explicit x.y.z.
# AUTOPUSH=1 pushes the tag, which is what triggers the release workflow.
release:
ifndef VERSION
	$(error VERSION is required; use VERSION=patch|minor|major|x.y.z [AUTOPUSH=1])
endif
	scripts/release/check-clean.sh
	go build ./...
	go vet ./...
	go test ./...
	VERSION="$(VERSION)" AUTOPUSH="$(AUTOPUSH)" scripts/release/tag.sh
