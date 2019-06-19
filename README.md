# terraform-provider-lightstep
Salad Bar 2019 Hackathon Project

* Install Terraform (will work with v0.11, not with v0.12)
* In `the terraform.tfstate` file, under resource, change the "nksingh-demo" values to your target project name, that you will also set in your `main.tf` file. Yes this is hacky but it's a hackathon project :)
  * The reason is that Terraform on an empty state will try to make a project but our project api create is not a public endpoint. Therefor, we're tricking Terrform into thinking the project already exists! Which it actually does. 
* In `main.tf` make sure you set the right organization that contains your target project, as well as check that your API key has the right permissions for that org/project as well
* Set LIGHTSTEP_API_KEY environment variable before the following

```bash

> export LIGHTSTEP_API_KEY = "[YOUR KEY HERE]"
> mkdir -p ~/.terraform.d/plugins
> go build -o ~/.terraform.d/plugins/terraform-provider-lightstep
> terraform init
> # write/edit your main.tf file schema
> terraform plan
> terraform apply
```

# TF File Schema

## Global

```terraform
provider "lightstep" {
  organization = "[ORG NAME]"
}
```

## Projects


```terraform
resource "lightstep_project" "[PROJECT]" {
  project_name = "[PROJECT]"
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

## Dashboards

```terraform
resource "lightstep_dashboard" "[DASHBOARD]" {
  project_name = "${lightstep_project.project.[PROJECT]}"
  dashboard_name = "[DASHBOARD NAME]"
  streams = [
    {
      stream_name = "${lightstep_stream.[STREAM].stream_name}"
      query = "${lightstep_stream.[STREAM].query}"
    }
  ]
}

```


## TODO

* Projects
  * Delete - Happens today but if new project is created with same name, crashes `app-staging`. Waiting on Pull Request
  * Update - What does it mean to update a project? What implications are there for historical data? 
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
* Test Terraform


