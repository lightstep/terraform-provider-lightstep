# Requires the provider to be built. Use 'make build' from the root makefile
.PHONY: generate-schema
generate-schema:
	terraform init -plugin-dir=../.terraform/providers
	terraform providers schema -json | sed 's#registry.terraform.io/lightstep/lightstep#lightstep#' > schema.json
