provider "lightstep" {
  api_key      = var.lightstep_api_key
  organization = var.lightstep_organization
}

terraform {
  required_providers {
    lightstep = {
      source  = "lightstep/lightstep"
      version = "~> 1.60.7"
    }
  }
  required_version = "~> 1.1.0"
}
