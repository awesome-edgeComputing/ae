.PHONY: build build-all clean package-tool

# Version
VERSION := 0.1.0

# Platforms
PLATFORMS := darwin/amd64 darwin/arm64 linux/amd64 linux/arm64

# Tools
GO := go
CARGO := cargo
PYTHON := python3
PYINSTALLER := pyinstaller

# Output directory
DIST_DIR := dist

PACKAGE_TOOL := scripts/package/package

build:
	./scripts/build.sh

package-tool:
	go build -o $(PACKAGE_TOOL) scripts/package/package.go

build-all: package-tool
	@for platform in $(PLATFORMS); do \
		echo "Building for $$platform..."; \
		os=$${platform%/*}; \
		arch=$${platform#*/}; \
		./scripts/build.sh $$os $$arch || exit 1; \
	done

clean:
	rm -rf dist
	rm -rf .venv
	rm -f $(PACKAGE_TOOL)
