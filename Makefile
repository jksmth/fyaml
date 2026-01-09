.DEFAULT_GOAL := help
SHELL = /bin/bash -o pipefail

.PHONY: build test lint clean install verify-examples verify-testdata all format format-check help venv hooks install-hooks uninstall-hooks clean-venv

venv ?= .venv
pip := $(venv)/bin/pip

$(pip): ## create venv using system python even when another venv is active
	@PATH=$${PATH#$${VIRTUAL_ENV}/bin:} python3 -m venv --clear $(venv)
	@$(pip) install --upgrade pip

$(venv): $(pip)
	@$(pip) install pre-commit~=4.0
	@touch $(venv)

help: ## display help message
	@awk 'BEGIN {FS = ":.*##"; printf "\nUsage:\n  make \033[36m<target> ...\033[0m\n\nTargets:\n"} /^[a-zA-Z_-]+:.*?##/ { printf "  \033[36m%-20s\033[0m %s\n", $$1, $$2 }' $(MAKEFILE_LIST)

build: ## build fyaml binary
	go build -o fyaml .
	@echo "Built fyaml"

test: ## run all tests
	go test -v ./...

test-coverage: ## run tests with coverage report
	go test -coverprofile=coverage.txt ./...
	go tool cover -html=coverage.txt -o coverage.html

lint: ## run golangci-lint
	@if command -v golangci-lint >/dev/null 2>&1; then \
		golangci-lint run; \
	else \
		echo "golangci-lint not installed, skipping. Install with: go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest"; \
	fi

clean: ## clean build artifacts
	rm -f fyaml fyaml.exe coverage.txt coverage.html
	rm -rf dist/ build/

clean-venv: ## delete the pre-commit venv
	rm -rf $(venv)

install: ## install fyaml via go install
	go install .

verify-examples: build ## verify example directories
	@echo "Verifying examples..."
	@./fyaml examples/basic > /dev/null && echo "✓ examples/basic" || (echo "✗ examples/basic failed" && exit 1)
	@./fyaml examples/with-at-files > /dev/null && echo "✓ examples/with-at-files" || (echo "✗ examples/with-at-files failed" && exit 1)
	@./fyaml examples/with-includes --enable-includes > /dev/null && echo "✓ examples/with-includes" || (echo "✗ examples/with-includes failed" && exit 1)

verify-testdata: verify-testdata-canonical verify-testdata-preserve ## verify testdata matches expected output for both modes

verify-testdata-canonical: build ## verify canonical mode testdata matches expected output
	@echo "Verifying canonical mode testdata..."
	@./fyaml testdata/simple/input --mode canonical -o testdata/simple/expected-canonical.yml --check && echo "✓ testdata/simple (canonical)" || (echo "✗ testdata/simple (canonical) failed" && exit 2)
	@./fyaml testdata/nested/input --mode canonical -o testdata/nested/expected-canonical.yml --check && echo "✓ testdata/nested (canonical)" || (echo "✗ testdata/nested (canonical) failed" && exit 2)
	@./fyaml testdata/at-root/input --mode canonical -o testdata/at-root/expected-canonical.yml --check && echo "✓ testdata/at-root (canonical)" || (echo "✗ testdata/at-root (canonical) failed" && exit 2)
	@./fyaml testdata/at-files/input --mode canonical -o testdata/at-files/expected-canonical.yml --check && echo "✓ testdata/at-files (canonical)" || (echo "✗ testdata/at-files (canonical) failed" && exit 2)
	@./fyaml testdata/ordering/input --mode canonical -o testdata/ordering/expected-canonical.yml --check && echo "✓ testdata/ordering (canonical)" || (echo "✗ testdata/ordering (canonical) failed" && exit 2)
	@./fyaml testdata/anchors/input --mode canonical -o testdata/anchors/expected-canonical.yml --check && echo "✓ testdata/anchors (canonical)" || (echo "✗ testdata/anchors (canonical) failed" && exit 2)
	@./fyaml testdata/includes/input --enable-includes --mode canonical -o testdata/includes/expected-canonical.yml --check && echo "✓ testdata/includes (canonical)" || (echo "✗ testdata/includes (canonical) failed" && exit 2)
	@./fyaml testdata/at-directories/input --mode canonical -o testdata/at-directories/expected-canonical.yml --check && echo "✓ testdata/at-directories (canonical)" || (echo "✗ testdata/at-directories (canonical) failed" && exit 2)

verify-testdata-preserve: build ## verify preserve mode testdata matches expected output
	@echo "Verifying preserve mode testdata..."
	@./fyaml testdata/simple/input --mode preserve -o testdata/simple/expected-preserve.yml --check && echo "✓ testdata/simple (preserve)" || (echo "✗ testdata/simple (preserve) failed" && exit 2)
	@./fyaml testdata/nested/input --mode preserve -o testdata/nested/expected-preserve.yml --check && echo "✓ testdata/nested (preserve)" || (echo "✗ testdata/nested (preserve) failed" && exit 2)
	@./fyaml testdata/at-root/input --mode preserve -o testdata/at-root/expected-preserve.yml --check && echo "✓ testdata/at-root (preserve)" || (echo "✗ testdata/at-root (preserve) failed" && exit 2)
	@./fyaml testdata/at-files/input --mode preserve -o testdata/at-files/expected-preserve.yml --check && echo "✓ testdata/at-files (preserve)" || (echo "✗ testdata/at-files (preserve) failed" && exit 2)
	@./fyaml testdata/ordering/input --mode preserve -o testdata/ordering/expected-preserve.yml --check && echo "✓ testdata/ordering (preserve)" || (echo "✗ testdata/ordering (preserve) failed" && exit 2)
	@./fyaml testdata/anchors/input --mode preserve -o testdata/anchors/expected-preserve.yml --check && echo "✓ testdata/anchors (preserve)" || (echo "✗ testdata/anchors (preserve) failed" && exit 2)
	@./fyaml testdata/includes/input --enable-includes --mode preserve -o testdata/includes/expected-preserve.yml --check && echo "✓ testdata/includes (preserve)" || (echo "✗ testdata/includes (preserve) failed" && exit 2)
	@./fyaml testdata/at-directories/input --mode preserve -o testdata/at-directories/expected-preserve.yml --check && echo "✓ testdata/at-directories (preserve)" || (echo "✗ testdata/at-directories (preserve) failed" && exit 2)

format: ## format Markdown, YAML, JSON, and Dockerfiles with dprint
	@if command -v dprint >/dev/null 2>&1; then \
		dprint fmt; \
	else \
		echo "dprint not installed. Install with: curl -fsSL https://dprint.dev/install.sh | sh"; \
	fi

format-check: ## check formatting without modifying files
	@if command -v dprint >/dev/null 2>&1; then \
		dprint check; \
	else \
		echo "dprint not installed. Install with: curl -fsSL https://dprint.dev/install.sh | sh"; \
	fi

all: clean build test lint ## clean, build, test, and lint

hooks: $(venv) ## run pre-commit git hooks on all files
	@$(venv)/bin/pre-commit run --show-diff-on-failure --all-files

install-hooks: $(venv) ## install pre-commit git hooks for repo
	@$(venv)/bin/pre-commit install -f --install-hooks
	@$(venv)/bin/pre-commit autoupdate

uninstall-hooks: $(venv) ## uninstall pre-commit git hooks for repo
	@$(venv)/bin/pre-commit clean
	@$(venv)/bin/pre-commit uninstall
