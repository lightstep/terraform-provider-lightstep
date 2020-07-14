.PHONY: build
build:
	go build -o terraform-provider-lightstep

.PHONY: install
install:
	terraform init

.PHONY: check_deps
check_deps:
	go mod tidy -v

.PHONY: test
test:
	go test ./...

.PHONY: fmt
fmt:
	go fmt

.PHONY: install-golangci-lint
install-golangci-lint:
	curl -sSfL https://raw.githubusercontent.com/golangci/golangci-lint/master/install.sh | sh -s v1.23.8

.PHONY: lint
lint:
	golangci-lint run --deadline 3m0s --no-config ./...