.PHONY: build
build:
	@go build -o terraform-provider-lightstep_v$(shell cat .go-version)
	@rm -f .terraform.lock.hcl
	@mkdir -p .terraform/providers/registry.terraform.io/lightstep/lightstep/$(shell cat .go-version)/$(shell uname -s | tr '[:upper:]' '[:lower:]')_amd64/
	@cp terraform-provider-lightstep_v$(shell cat .go-version) .terraform/providers/registry.terraform.io/lightstep/lightstep/$(shell cat .go-version)/$(shell uname -s | tr '[:upper:]' '[:lower:]')_amd64/terraform-provider-lightstep_v$(shell cat .go-version)
	@mkdir -p terraform.d/plugins/terraform.lightstep.com/lightstep-org/lightstep/$(shell cat .go-version)/$(shell uname -s | tr '[:upper:]' '[:lower:]')_amd64/
	@cp terraform-provider-lightstep_v$(shell cat .go-version) terraform.d/plugins/terraform.lightstep.com/lightstep-org/lightstep/$(shell cat .go-version)/$(shell uname -s | tr '[:upper:]' '[:lower:]')_amd64/terraform-provider-lightstep

.PHONY: install
install:
	@terraform init

.PHONY: check-deps
check_deps:
	@go mod tidy -v

.PHONY: test
test:
	@go test ./...

.PHONY: fmt
fmt:
	@go fmt

.PHONY: install-golangci-lint
install-golangci-lint:
	@curl -sSfL https://raw.githubusercontent.com/golangci/golangci-lint/master/install.sh | sh -s v1.23.8

.PHONY: lint
lint:
	@golangci-lint run --deadline 3m0s --no-config ./...

.PHONY: acc-test
acc-test:
	@TF_ACC=true LIGHTSTEP_API_KEY=${LIGHTSTEP_API_KEY_PUBLIC} LIGHTSTEP_ORG="LightStep" LIGHTSTEP_ENV="public" go test -v ./lightstep

.PHONY: ensure-clean-repo
ensure-clean-repo:
	@scripts/ensure_clean_repo.sh
