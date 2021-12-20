# Lightstep Terraform Provider

-   Website: https://www.terraform.io
-   [![Gitter chat](https://badges.gitter.im/hashicorp-terraform/Lobby.png)](https://gitter.im/hashicorp-terraform/Lobby)
-   Mailing list: [Google Groups](http://groups.google.com/group/terraform-tool)

## Requirements

-   [Terraform](https://www.terraform.io/downloads.html) 1.x

## Using the provider

* Install Terraform (v1.0)
* Build the binary & initialize
```
make build
tf init
```

* Configure a provider block (see main.tf for an example)
```
provider "lightstep" {
  environment = public
  api_key_env_var = (environment variable where your LS API key is stored)
  organization = (organization name)
}
```

## Testing the provider

The integration tests create, update, and destroy real resources in a dedicated project in public:
[terraform-provider-tests](https://app.lightstep.com/terraform-provider-tests/service-directory/android/deployments)

To run the tests, first get an API key with a Member role for the public environment and run:
```
LIGHTSTEP_API_KEY_PUBLIC=(your api key here) make acc-test
```

