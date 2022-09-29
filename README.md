# Lightstep Terraform Provider

⚠️ This provider is still under active development and we're working on adding new features and functionality.

- Website: https://www.terraform.io
- [![Gitter chat](https://badges.gitter.im/hashicorp-terraform/Lobby.png)](https://gitter.im/hashicorp-terraform/Lobby)
- Mailing list: [Google Groups](http://groups.google.com/group/terraform-tool)
- [Hashicorp Terraform Registry](https://registry.terraform.io/providers/lightstep/lightstep/latest)

## Requirements

- [Terraform](https://www.terraform.io/downloads.html) v1.x+

## Using the provider

- [Install Terraform](https://www.terraform.io/downloads)
- Write some code to initialize the provider with your Lightstep organization and [API Key](https://docs.lightstep.com/docs/create-and-manage-api-keys) with `member` permissions:

```
terraform {
  required_providers {
    lightstep = {
      source = "lightstep/lightstep"
      version = "1.61.1"
    }
  }
}

provider "lightstep" {
  api_key         = "your-lightstep-org-api-key"
  organization    = "your-lightstep-organization"
}

# Example: Create AWS EC2 Dashboard
module "aws-dashboards" {
  source            = "lightstep/aws-dashboards/lightstep//modules/ec2-dashboard"
  lightstep_project = "your-lightstep-project"
}
```

- Run `terraform init`
- Add some code to define dashboards, streams, alerts, and more. [See documentation](https://registry.terraform.io/providers/lightstep/lightstep/latest/docs) for examples or use [pre-built Lightstep Terraform modules](https://registry.terraform.io/namespaces/lightstep).
- After setting an environment variable with your API Key that matches the name in the provider configuration above, run `terraform plan` to preview changes.

:warning: If you're creating many Lightstep resources at once, we recommend running the `apply` with the `parallelism` flag set to a low value to avoid API 500 errors:

```
   # Avoids 500 errors when creating many resources.
   terraform apply -parallelism=1
```

## Development

See [`DEVELOPMENT.md`](DEVELOPMENT.md).

## Exporter

It's possible to export an existing Lightstep dashboard to HCL code using the provider. This allows you to generate terrform code for a dashboard you created in the Lightstep UI ("reverse terraform").

The `exporter` utility is built-in to the provider binary and requires certain environment variables to be set:

```
$ export LIGHTSTEP_API_KEY=....
$ export LIGHTSTEP_ORG=your-org
$ export LIGHTSTEP_ENV=public

# exports to console dashboard id = rZbPJ33q from project terraform-shop
$ go run github.com/lightstep/terraform-provider-lightstep exporter dashboard terraform-shop rZbPJ33q
```
