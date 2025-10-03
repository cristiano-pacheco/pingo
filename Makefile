# Check to see if we can use ash, in Alpine images, or default to BASH.
SHELL_PATH = /bin/ash
SHELL = $(if $(wildcard $(SHELL_PATH)),/bin/ash,/bin/bash)

# Load environment variables from .env file if it exists
ifneq (,$(wildcard ./.env))
    include .env
    export
endif

.PHONY: all
all: install-libs lint test cover

# ==============================================================================
# Install dependencies

.PHONY: install-libs
install-libs:
	go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest
	go install golang.org/x/vuln/cmd/govulncheck@latest
	go install github.com/vektra/mockery/v2@latest
	go install github.com/swaggo/swag/cmd/swag@latest
	go install go.uber.org/nilaway/cmd/nilaway@latest

# ==============================================================================
# Administration

.PHONY: run
run:
	go run ./main.go server

.PHONY: migrate
migrate:
	go run ./main.go db:migrate

# ==============================================================================
# Running tests within the local computer

.PHONY: static
static: lint vuln-check nilaway

.PHONY: lint
lint:
	golangci-lint run ./... --allow-parallel-runners

.PHONY: vuln-check
vuln-check:
	govulncheck -show verbose ./... 

.PHONY: nilaway
nilaway:
	nilaway --include-pkgs="github.com/cristiano-pacheco/pingo" --exclude-pkgs="vendor/" ./...

.PHONY: test
test:
	CGO_ENABLED=0 go test ./...

.PHONY: test-integration
test-integration:
	APP_BASE_URL=http://localhost:9000 CGO_ENABLED=0 go test -v -race -timeout=30s -tags=integration ./test/integration/...

.PHONY: cover
cover:
	mkdir -p reports
	go test -race -coverprofile=reports/cover.out -coverpkg=./... ./... && \
	go tool cover -html=reports/cover.out -o reports/cover.html

.PHONY: update-mocks
update-mocks:
	mockery

.PHONY: update-swagger
update-swagger:
	go install github.com/swaggo/swag/cmd/swag@latest
	swag i --parseDependency

# ==============================================================================
# NOTES
#
# RSA Keys
# 	To generate a private key PEM file.
# 	$ openssl genpkey -algorithm RSA -pkeyopt rsa_keygen_bits:2048 | base64 | tr -d '\n' > private_key_base64.txt
#
#	To convert the txt file to a PEM file.
#   base64 -D -i private_key_base64.txt -o private.pem