provider "lightstep" {
  api_key_env_var = "LIGHTSTEP_API_KEY"
  organization    = var.lightstep_organization
}

terraform {
  required_providers {
    lightstep = {
      source  = "lightstep/lightstep"
      version = "1.51.6"
    }
  }
  required_version = "~> 1.1.0"
}