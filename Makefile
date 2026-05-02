# Variables
GOHOSTOS:=$(shell go env GOHOSTOS)
GOHOSTARCH:=$(shell go env GOHOSTARCH)
GOPATH:=$(shell go env GOPATH)
VERSION=$(shell git describe --tags --always 2>/dev/null || echo unknown)
BUILD_TIME=$(shell date -u +%Y-%m-%dT%H:%M:%SZ)
GIT_COMMIT=$(shell git rev-parse --short HEAD 2>/dev/null || echo unknown)

# Inject build metadata into the binary so /hello can return the real commit/time.
# 这两个变量必须是 var（不能是 const），见 internal/version/version.go。
VERSION_PKG=github.com/lin-snow/ech0/internal/version
LDFLAGS=-X $(VERSION_PKG).Commit=$(GIT_COMMIT) -X $(VERSION_PKG).BuildTime=$(BUILD_TIME)

# Docker variables
DOCKER_REGISTRY?=sn0wl1n
IMAGE_NAME?=ech0
IMAGE_TAG?=latest
OS?=$(if $(GOHOSTOS),$(GOHOSTOS),linux)
ARCH?=$(if $(GOHOSTARCH),$(GOHOSTARCH),amd64)

.PHONY: help air-install run dev web-dev check dev-lint lint fmt test wire wire-check swagger spdx spdx-check build bump build-image push-image

# Semver pattern: X.Y.Z, optionally followed by a -prerelease suffix.
# (escape $ so make doesn't expand, then escape $$ to a literal $ in shell)
SEMVER_PATTERN := ^[0-9]+\.[0-9]+\.[0-9]+(-[0-9A-Za-z.-]+)?$$

AIR_BIN := $(shell command -v air 2>/dev/null || echo "$(GOPATH)/bin/air")

help:
	@echo "Available targets:"
	@echo "  make run         - Run backend in serve mode"
	@echo "  make dev         - Run backend with Air hot reload"
	@echo "  make air-install - Install Air to GOPATH/bin"
	@echo "  make web-dev     - Run frontend dev server"
	@echo "  make build       - Build local binary with version/commit injected"
	@echo "  make bump NEW_VERSION=X.Y.Z"
	@echo "                   - Bump internal/version.Version + sanity-check (does NOT commit/tag)"
	@echo "  make check       - Backend fmt/lint + web format/lint + i18n checks"
	@echo "  make dev-lint    - Backend fmt/lint + web format/lint + i18n checks"
	@echo "  make lint        - Run golangci-lint checks"
	@echo "  make fmt         - Run golangci-lint formatters"
	@echo "  make test        - Run Go tests"
	@echo "  make wire        - Generate DI code via Wire"
	@echo "  make wire-check  - Verify Wire code is up-to-date"
	@echo "  make swagger     - Regenerate Swagger docs"
	@echo "  make spdx        - Add SPDX/Copyright headers to new .go/.ts/.vue files"
	@echo "  make spdx-check  - Fail if any source file is missing the SPDX header"
	@echo "  make build-image - Build Docker image"
	@echo "  make push-image  - Push Docker image"

air-install:
	go install github.com/air-verse/air@latest

run:
	go run -ldflags "$(LDFLAGS)" ./cmd/ech0 serve

build:
	go build -ldflags "$(LDFLAGS)" -o ./bin/ech0 ./cmd/ech0

# Prepare a clean version-bump commit. This target only EDITS files —
# it never auto-commits, never tags, never pushes. The next-step commands
# are printed for the developer to run after eyeballing the diff.
# See docs/dev/release-process.md for the full procedure.
bump:
	@if [ -z "$(NEW_VERSION)" ]; then \
		echo "✘ Usage: make bump NEW_VERSION=X.Y.Z"; \
		echo "  e.g. make bump NEW_VERSION=4.6.5"; \
		exit 1; \
	fi
	@echo "$(NEW_VERSION)" | grep -Eq '$(SEMVER_PATTERN)' \
		|| { echo "✘ '$(NEW_VERSION)' is not valid semver (expected X.Y.Z[-prerelease])"; exit 1; }
	@if [ -n "$$(git status --porcelain)" ]; then \
		echo "✘ Working tree dirty — commit or stash first so the bump commit is clean."; \
		git status --short; \
		exit 1; \
	fi
	@OLD_VERSION="$$(grep -E '^[[:space:]]*Version[[:space:]]*=[[:space:]]*\"' internal/version/version.go \
	                  | head -n1 \
	                  | sed -E 's/.*\"([^\"]+)\".*/\1/')"; \
	if [ -z "$$OLD_VERSION" ]; then \
		echo "✘ Could not extract current Version from internal/version/version.go"; \
		exit 1; \
	fi; \
	if [ "$$OLD_VERSION" = "$(NEW_VERSION)" ]; then \
		echo "✘ Version is already $$OLD_VERSION — nothing to bump."; \
		exit 1; \
	fi; \
	echo "→ bumping $$OLD_VERSION → $(NEW_VERSION)"; \
	sed -i.bak -E 's/^([[:space:]]*Version[[:space:]]*=[[:space:]]*\")[^\"]+(\")/\1$(NEW_VERSION)\2/' internal/version/version.go; \
	rm -f internal/version/version.go.bak
	@echo "→ verifying go build still succeeds..."
	@go build ./... >/dev/null || { echo "✘ go build failed after bump — reverting"; git checkout -- internal/version/version.go; exit 1; }
	@echo ""
	@echo "✓ Version bumped. Diff:"
	@git --no-pager diff -- internal/version/version.go
	@echo ""
	@echo "Next steps (review the diff above, then run):"
	@echo ""
	@echo "  # 1. Update CHANGELOG.md: rename [Unreleased] → [$(NEW_VERSION)] - $$(date -u +%Y-%m-%d), open a new empty [Unreleased]"
	@echo "  # 2. Commit + tag:"
	@echo "       git commit -am 'chore(release): v$(NEW_VERSION)'"
	@echo "       git tag -a v$(NEW_VERSION) -m 'Release v$(NEW_VERSION)'"
	@echo "  # 3. Push to trigger release workflow:"
	@echo "       git push origin main"
	@echo "       git push origin v$(NEW_VERSION)"

dev:
	@if [ ! -x "$(AIR_BIN)" ]; then \
		echo "air not found, installing..."; \
		$(MAKE) air-install; \
	fi
	"$(AIR_BIN)" -c .air.toml

web-dev:
	cd web && pnpm dev

check:
	$(MAKE) dev-lint

dev-lint:
	@bash scripts/check.sh

lint:
	golangci-lint run

fmt:
	golangci-lint fmt

test:
	go test ./...

wire:
	go generate ./internal/di

wire-check: wire
	git diff --exit-code -- internal/di/wire_gen.go

swagger:
	swag init -g internal/server/server.go -o internal/swagger --parseInternal

spdx:
	node scripts/add-spdx-headers.mjs

spdx-check:
	node scripts/add-spdx-headers.mjs --check

build-image:
	@echo "Building image for platform: $(OS)/$(ARCH)"
	docker build --platform $(OS)/$(ARCH) \
		--build-arg TARGETOS=$(OS) \
		--build-arg TARGETARCH=$(ARCH) \
		--build-arg GIT_COMMIT=$(GIT_COMMIT) \
		--build-arg BUILD_TIME=$(BUILD_TIME) \
		-t $(DOCKER_REGISTRY)/$(IMAGE_NAME):$(IMAGE_TAG) -f docker/build.Dockerfile .
push-image:
	docker push $(DOCKER_REGISTRY)/$(IMAGE_NAME):$(IMAGE_TAG)

