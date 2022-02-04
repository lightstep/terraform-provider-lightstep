variable "project" {
  description = "Name of Lightstep project"
  default     = "terraform-provider-tests"
  type        = string
}

variable "lightstep_organization" {
  description = "Name of Lightstep organization"
  type        = string
}
