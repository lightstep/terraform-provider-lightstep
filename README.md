# terraform-provider-lightstep
Salad Bar 2019 Hackathon Project

* Set LIGHTSTEP_API_KEY environment variable before the following

* `mkdir -p ~/.terraform.d/plugins`
* `go build -o ~/.terraform.d/plugins/terraform-provider-lightstep`
* `terraform init`
* `terraform plan`
* `terraform apply`


# TF File Schema

## Global

```HCL
provider "lightstep" {
  organization = "[ORG NAME]"
}
```

## Projects

```HCL
resource "lightstep_project" "[PROJECT]" {
  project_name = "[PROJECT]"
}
```

## Dashboards

```HCL

```

## Streams

```HCL
resource "lightstep_stream" "[STREAM]" {
  project_name = "${lightstep_project.project.[PROJECT]}"
  stream_name = "[STREAM]"
  query = "[QUERY]"
  depends_on = ["lightstep_project.[PROJECT]"]
}
```

