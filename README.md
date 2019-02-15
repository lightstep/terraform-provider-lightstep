# terraform-provider-lightstep
Salad Bar 2019 Hackathon Project

* Set LIGHTSTEP_API_KEY environment variable before the following

```bash

> export LIGHTSTEP_API_KEY = "[YOUR KEY HERE]"
> mkdir -p ~/.terraform.d/plugins
> go build -o ~/.terraform.d/plugins/terraform-provider-lightstep
> terraform init
> terraform plan
> terraform apply

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

## Dashboards

```terraform

```

## Streams

```terraform
resource "lightstep_stream" "[STREAM]" {
  project_name = "${lightstep_project.project.[PROJECT]}"
  stream_name = "[STREAM]"
  query = "[QUERY]"
  depends_on = ["lightstep_project.[PROJECT]"]
}