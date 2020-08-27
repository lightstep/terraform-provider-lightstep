# terraform-provider-lightstep

## Getting Started
* Install Terraform (v0.12)
* Build the binary & initialize
```
make build
tf init
```

* Configure a provider block (see main.tf)
```
provider "lightstep" {
  environment = (staging|meta|public)
  api_key_env_var = (environment variable where your LS api key is stored)
  organization = (organization name)
}
```

## Repo Overview
* **.go-version** - the current version of the terraform provider
* **provider.go** - top level set up for the provider itself, specifies what resources it can manage
* **resource_X.go** - set up for each resource specifying how to parse the block from terraform and CRUD/import methods
* **resource_X_test.go** - acceptance test for the resource that will exercise the full functionality creating/modifying/deleting actual resources in the lightstep-terraform-provider project in LightStep-Integration in public.
* **lightstep/** - client to communicate to LS 
* **main.tf** - an terraform config file with examples for every resource
* **scripts/** - catchall dir to hold scripts, currently holds a script needed for a CI check


## Testing
The integration tests create, update, and destroy real resources in a dedicated project in public:
[terraform-provider-tests](https://app.lightstep.com/terraform-provider-tests/service-directory/android/deployments)

To run the tests, first get an API key with a Member role for the public environment and run:
```
LIGHTSTEP_API_KEY=(your api key here) make acc-test
```


## TODO 
* Streams 
  * Update - What does it mean to change the stream query in the TF file? Is the previous stream no longer required and should be deleted (AS-IS)?
* Dashboards
  * Currently required to pass in verbose stream parameters to associate
  * Terraform 0.12 should have ability to do for loops in TF file, potentially making it easier. 
  * API requires stream name and query to associate with dashboard. This is because the API makes new streams if they don't exist which is problematic from Terraform's perspective because that is not reflected in the state  Change this to just accept stream IDs in the SDK.
* API
  * Rate limiting - when resources are created in parallel, might run into rate limits
  * 500 error for lock when streams are being created
* SDK - Separate into a repo



