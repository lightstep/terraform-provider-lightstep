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

* Updating Projects - what does this mean?
* Deleting Projects and creating new one with same name - waiting on pull request
* Only passing in Stream IDs/TF aliases for Dashboards instead of duplicating data
* Rate Limiting APIs

