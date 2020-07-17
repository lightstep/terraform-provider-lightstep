# terraform-provider-lightstep

## Getting Started
* Install Terraform (v0.12)
* Set LIGHTSTEP_API_KEY & LIGHTSTEP_ORG environment variables
* Set your project name in the variable block of `main.tf`
* Build the binary 
```bash

> mkdir -p ~/.terraform.d/plugins
> go build -o ~/.terraform.d/plugins/terraform-provider-lightstep
> terraform init
> # write/edit your main.tf file schema
> terraform plan
> terraform apply
```

# TF File Schema

## Global
Setting the provider block is optional since you can set LIGHTSTEP_ORG & LIGHTSTEP_API_KEY 
as environment variables.

```terraform
provider "lightstep" {
  organization = "[ORG NAME]"
  api_key      = "[API KEY]"
  
}
```

## Streams

```terraform
resource "lightstep_stream" "[STREAM]" {
  project_name = "${lightstep_project.project.[PROJECT]}"
  stream_name = "[STREAM]"
  query = "[QUERY]"
}
```

To import a stream:
`terraform import lightstep_stream.<resource name> project.streamID`

## Dashboards

```terraform
resource "lightstep_dashboard" "[DASHBOARD]" {
  project_name = "${lightstep_project.project.[PROJECT]}"
  dashboard_name = "[DASHBOARD NAME]"
  stream_ids = [STREAM_IDS]

}

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


