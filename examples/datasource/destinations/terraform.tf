terraform {
  required_providers {
    lightstep = {
     source = "lightstep/lightstep"
     version = "~> 1.81.0"
    }
  }
}

provider "lightstep" {
  api_key      = var.lightstep_api_key
  organization = var.lightstep_organization
}