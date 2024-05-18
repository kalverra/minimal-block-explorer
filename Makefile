test:
	set -euo pipefail
	LOG_LEVEL="warn" go test -coverprofile cover.out -json -v ./... 2>&1 | tee /tmp/gotest.log | gotestfmt

race:
	go test -race  ./...

lint:
	golangci-lint --color=always run ./... --fix -v

install:
	pre-commit install
