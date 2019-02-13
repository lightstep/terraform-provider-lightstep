provider "lightstep" {
  # api_key = "${var.lighstep_api_key}"
  # api_url = "${var.lightstep_host}"
  organization = "test_org"
}

resource "lightstep_stream" "test_stream" {
  project = "test_project"
  name = "test_name"
  query = "test_query"
  stream_id = "test_stream_id"
}