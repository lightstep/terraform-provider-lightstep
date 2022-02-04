## Provider Development

## Testing the provider

If you're contributing changes or code to the provider, the integration tests create, update, and destroy real resources in a Lightstep-managed integration environment.

To run the tests, first get an API key with a Member role for the public environment and run:
```
LIGHTSTEP_API_KEY_PUBLIC=(your api key here) make acc-test
```

## Using a local build for development (vs the one in the registry)

1) Update the version in `.version` and run `make build`
2) Create a [`.terraformrc`](https://www.terraform.io/cli/config/config-file) in your $HOME directory that points this checked-out repository with your changes.

```
provider_installation {  
    filesystem_mirror {    
        path    = "/Users/your-username/workspace/terraform-provider-lightstep/.terraform/providers"    
        include = ["registry.terraform.io/lightstep/lightstep"]  
    }
    direct {
        exclude = ["lightstep/lightstep"]
    }
}
```

3) Import the module in another project pinned to the version that matches `.version`. Delete any terraform lock files and run `terraform init`. 


## Exporter (experimental)

It's possible to export an existing dashboard to conformant HCL code using the provider. The `exporter` utility is built-in to the provider binary and requires certain environment variables to be set:

```
$ export LIGHTSTEP_API_KEY=....
$ export LIGHTSTEP_ORG=your-org
# exports to console dashboard id = rZbPJ33q from project terraform-shop
$ .terraform/providers/registry.terraform.io/lightstep/lightstep/1.51.5/darwin_amd64/terraform-provider-lightstep_v1.51.5 exporter dashboard terraform-shop rZbPJ33q
```