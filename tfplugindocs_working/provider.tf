provider "lightstep" {
}

terraform {
  required_providers {
    lightstep = {
      source  = "registry.terraform.io/lightstep/lightstep"
    }
  }
}