# Currently always builds for amd64
#PLATFORM_ARCH=$(shell uname -m | tr '[:upper:]' '[:lower:]')
PLATFORM_ARCH=amd64
PLATFORM_NAME=$(shell uname -s | tr '[:upper:]' '[:lower:]')
VERSION_TAG=$(shell cat .go-version)

.PHONY: build
build:
	GOARCH=$(PLATFORM_ARCH) go build -o terraform-provider-lightstep_v$(VERSION_TAG)
	@rm -f .terraform.lock.hcl
	@mkdir -p .terraform/providers/registry.terraform.io/lightstep/lightstep/$(VERSION_TAG)/$(PLATFORM_NAME)_$(PLATFORM_ARCH)/
	@cp terraform-provider-lightstep_v$(VERSION_TAG) .terraform/providers/registry.terraform.io/lightstep/lightstep/$(VERSION_TAG)/$(PLATFORM_NAME)_$(PLATFORM_ARCH)/terraform-provider-lightstep_v$(VERSION_TAG)
	@mkdir -p terraform.d/plugins/terraform.lightstep.com/lightstep-org/lightstep/$(VERSION_TAG)/$(PLATFORM_NAME)_$(PLATFORM_ARCH)/
	@cp terraform-provider-lightstep_v$(VERSION_TAG) terraform.d/plugins/terraform.lightstep.com/lightstep-org/lightstep/$(VERSION_TAG)/$(PLATFORM_NAME)_$(PLATFORM_ARCH)/terraform-provider-lightstep

.PHONY: install
install:
	terraform init

.PHONY: check-deps
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

.PHONY: acc-test
acc-test:
ifndef LIGHTSTEP_API_KEY_PUBLIC
	$(error LIGHTSTEP_API_KEY_PUBLIC must be defined for acc-test)
endif
	@TF_ACC=true LIGHTSTEP_API_KEY=${LIGHTSTEP_API_KEY_PUBLIC} LIGHTSTEP_ORG="LightStep" LIGHTSTEP_ENV="public" go test -v ./lightstep

test-local:
	@TF_ACC=true LIGHTSTEP_API_BASE_URL=http://localhost:11000 LIGHTSTEP_API_KEY=${LIGHTSTEP_LOCAL_API_KEY} LIGHTSTEP_ORG="LightStep" LIGHTSTEP_ENV="public" go test -v ./lightstep 

test-staging:
	@TF_ACC=true LIGHTSTEP_API_BASE_URL=http://api-staging.lightstep.com LIGHTSTEP_API_KEY=${LIGHTSTEP_STAGING_API_KEY} LIGHTSTEP_ORG="LightStep" LIGHTSTEP_ENV="staging" go test -v ./lightstep 

.PHONY: ensure-clean-repo
ensure-clean-repo:
	scripts/ensure_clean_repo.sh

.PHONY: clean
clean:
	-rm -rf terraform.d .terraform
	-rm terraform-provider-lightstep_v*
