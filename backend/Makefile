.PHONY: build generate test test-unit test-integration test-growth-contract test-e2e

VERSION ?= $(shell ./scripts/resolve-version.sh)
UPSTREAM_VERSION ?= $(shell tr -d '\r\n' < ./cmd/server/UPSTREAM_VERSION)
LDFLAGS ?= -s -w -X main.Version=$(VERSION) -X main.UpstreamVersion=$(UPSTREAM_VERSION)

build: validate-version
	CGO_ENABLED=0 go build -ldflags="$(LDFLAGS)" -trimpath -o bin/server ./cmd/server

validate-version:
	./scripts/validate-version.sh

generate:
	go generate ./ent
	go generate ./cmd/server

test:
	go test ./...
	golangci-lint run ./...

test-unit:
	go test -tags=unit ./...

test-integration:
	go test -tags=integration ./...

test-growth-contract:
	go test -tags=unit ./internal/service ./internal/server/routes ./internal/server/middleware \
		-run 'GrowthRegistration|GrowthReferral|AuthService_Register_RecordsGrowth' -count=1

test-e2e:
	./scripts/e2e-test.sh

test-e2e-local:
	go test -tags=e2e -v -timeout=300s ./internal/integration/...
