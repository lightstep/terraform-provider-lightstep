# terraform-provider-lightstep
Salad Bar 2019 Hackathon Project

* Set LIGHTSTEP_API_KEY environment variable before the following
* `go build -o terraform-provider-lightstep`
* `terraform init`
* `terraform apply`

@cody: Check out the `resource_stream.go` and `resource_project.go` files for example on how to do CREATE. 

`main.tf` has an example schema which when you apply, shows up in our test project on staging. You can also create new projects (just change the name) but probably wait until Julian can make his fix that stops staging env from breaking