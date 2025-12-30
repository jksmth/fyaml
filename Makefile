.PHONY: build test lint clean install verify-examples verify-testdata all

build:
	go build -o fyaml .
	@echo "Built fyaml"

test:
	go test -v ./...

test-coverage:
	go test -coverprofile=coverage.txt ./...
	go tool cover -html=coverage.txt -o coverage.html

lint:
	@if command -v golangci-lint >/dev/null 2>&1; then \
		golangci-lint run; \
	else \
		echo "golangci-lint not installed, skipping. Install with: go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest"; \
	fi

clean:
	rm -f fyaml fyaml.exe coverage.txt coverage.html
	rm -rf dist/ build/

install:
	go install .

verify-examples: build
	@echo "Verifying examples..."
	@./fyaml pack examples/basic > /dev/null && echo "✓ examples/basic" || (echo "✗ examples/basic failed" && exit 1)
	@./fyaml pack examples/with-at-files > /dev/null && echo "✓ examples/with-at-files" || (echo "✗ examples/with-at-files failed" && exit 1)

verify-testdata: build
	@echo "Verifying testdata..."
	@./fyaml pack testdata/simple/input -o testdata/simple/expected.yml --check && echo "✓ testdata/simple" || (echo "✗ testdata/simple failed" && exit 2)
	@./fyaml pack testdata/nested/input -o testdata/nested/expected.yml --check && echo "✓ testdata/nested" || (echo "✗ testdata/nested failed" && exit 2)
	@./fyaml pack testdata/at-root/input -o testdata/at-root/expected.yml --check && echo "✓ testdata/at-root" || (echo "✗ testdata/at-root failed" && exit 2)
	@./fyaml pack testdata/at-files/input -o testdata/at-files/expected.yml --check && echo "✓ testdata/at-files" || (echo "✗ testdata/at-files failed" && exit 2)
	@./fyaml pack testdata/ordering/input -o testdata/ordering/expected.yml --check && echo "✓ testdata/ordering" || (echo "✗ testdata/ordering failed" && exit 2)

all: clean build test lint

