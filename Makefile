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

verify-testdata: build ## verify testdata matches expected output
	@echo "Verifying testdata..."
	@./fyaml testdata/simple/input -o testdata/simple/expected.yml --check && echo "✓ testdata/simple" || (echo "✗ testdata/simple failed" && exit 2)
	@./fyaml testdata/nested/input -o testdata/nested/expected.yml --check && echo "✓ testdata/nested" || (echo "✗ testdata/nested failed" && exit 2)
	@./fyaml testdata/at-root/input -o testdata/at-root/expected.yml --check && echo "✓ testdata/at-root" || (echo "✗ testdata/at-root failed" && exit 2)
	@./fyaml testdata/at-files/input -o testdata/at-files/expected.yml --check && echo "✓ testdata/at-files" || (echo "✗ testdata/at-files failed" && exit 2)
	@./fyaml testdata/ordering/input -o testdata/ordering/expected.yml --check && echo "✓ testdata/ordering" || (echo "✗ testdata/ordering failed" && exit 2)
	@./fyaml testdata/anchors/input -o testdata/anchors/expected.yml --check && echo "✓ testdata/anchors" || (echo "✗ testdata/anchors failed" && exit 2)
	@./fyaml testdata/includes/input --enable-includes -o testdata/includes/expected.yml --check && echo "✓ testdata/includes" || (echo "✗ testdata/includes failed" && exit 2)
	@./fyaml testdata/at-directories/input -o testdata/at-directories/expected.yml --check && echo "✓ testdata/at-directories" || (echo "✗ testdata/at-directories failed" && exit 2)

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
