# Lightstep Terraform Provider

⚠️ *Beta* This provider is still under active development and we're working on adding new features and functionality.

-   Website: https://www.terraform.io
-   [![Gitter chat](https://badges.gitter.im/hashicorp-terraform/Lobby.png)](https://gitter.im/hashicorp-terraform/Lobby)
-   Mailing list: [Google Groups](http://groups.google.com/group/terraform-tool)
-   [Hashicorp Terraform Registry](https://registry.terraform.io/providers/lightstep/lightstep/latest)

## Requirements

-   [Terraform](https://www.terraform.io/downloads.html) v1.1.x

## Using the provider

* [Install Terraform](https://www.terraform.io/downloads)
* Write some code to initialize the provider with your Lightstep organization and [API Key](https://docs.lightstep.com/docs/create-and-manage-api-keys) with `member` permissions:
```
terraform {
  required_providers {
    lightstep = {
      source = "lightstep/lightstep"
      version = "1.51.1"
    }
  }
}

provider "lightstep" {
  environment     = "public"
  api_key_env_var = "LIGHTSTEP_API_KEY_PUBLIC"
  organization    = "LightStep"
}
```
* Run `terraform init`
* Add some code to define dashboards, streams, alerts, and more. See documentation for examples.
* After setting an environment variable with your API Key that matches the name in the provider configuration above, run `terraform plan` to preview changes.

## Testing the provider (development only)

If you're contributing changes or code to the provider, the integration tests create, update, and destroy real resources in a Lightstep-managed integration environment.

To run the tests, first get an API key with a Member role for the public environment and run:
```
LIGHTSTEP_API_KEY_PUBLIC=(your api key here) make acc-test
```
