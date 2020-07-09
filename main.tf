provider "lightstep" {
  organization = "YOUR ORG HERE"
  api_key = "YOUR API KEY HERE"
}

variable "project" {
  type = string
  default = "YOUR PROJECT HERE"
}

#############################################################
# Streams
#############################################################
resource "lightstep_stream" "stream_5" {
  project_name = var.project
  stream_name = "BEEMO! Charge! UPDATED NAME"
  query = "operation IN (\"api/v1/charge\") AND \"customer_id\" NOT IN (\"BEEMO\")"
}

resource "lightstep_stream" "stream_3" {
  project_name = var.project
  stream_name = "Non BEEMO Charges All Else"
  query = "operation IN (\"api/v1/charge\") AND \"customer_id\" NOT IN (\"BEEMO\", \"oops\")"
}

#############################################################
# Dashboards
#############################################################

resource "lightstep_dashboard" "test" {
  project_name = var.project
  dashboard_name = "terrrrrraform"
  stream_ids = [lightstep_stream.stream_5.id, lightstep_stream.stream_3.id]
}




