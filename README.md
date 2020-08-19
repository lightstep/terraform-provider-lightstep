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
  project_name = "${var.project}"
  stream_name = "STREAM NAME"
  query = "QUERY"
}
```

To import a stream:
`terraform import lightstep_stream.<resource name> project.streamID`

## Dashboards

```terraform
resource "lightstep_dashboard" "[DASHBOARD]" {
  project_name = "${var.project}"
  dashboard_name = "DASHBOARD NAME"
  stream_ids = ["STREAM_IDS"]

}

```

To import a dashboard:
`terraform import lightstep_stream.<resource name> project.dashboardID`

## Conditions

```
resource "lightstep_condition" "beemo_errors" {
  project_name = var.project
  condition_name = "CONDITION NAME"
  expression = "EXPRESSION"
  evaluation_window_ms = MS
  stream_id = "STREAM_ID"
}
```

To import a condition:
`terraform import lightstep_stream.<resource name> project.conditionID`

## Destinations

#### Webhook Destination

```
resource "lightstep_webhook_destination" "my_destination" {
     project_name = var.project
     destination_name = "NAME OF DESTINATION"
     url = "https://www.YOUR-URL.net"
     destination_type = "webhook"
     custom_headers = {
       "Access-Control-Max-Age" = "120"
     }
   }
```

To import a destination:
`terraform import lightstep_destination.<resource_name> <project>.destinationID`


## Testing
The integration tests create, update, and destroy real resources in a dedicated project in public terraform-provider-tests
[project link](https://app.lightstep.com/terraform-provider-tests/service-directory/android/deployments)

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



