provider "lightstep" {
  api_key      = var.lightstep_api_key
  organization = var.lightstep_organization
}

terraform {
  required_providers {
    lightstep = {
      source  = "lightstep/lightstep"
      version = "~> 1.61.1"
    }
  }
  required_version = "~> 1.1.0"
}
