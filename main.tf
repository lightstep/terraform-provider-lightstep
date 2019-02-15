provider "lightstep" {
  organization = "LightStep"
}

resource "lightstep_stream" "stream_1" {
  project_name = "${lightstep_project.project.project_name}"
  stream_name = "test_stream_1"
  query = "tag:\"error\"=\"true\""
  depends_on = ["lightstep_project.project"]
}

resource "lightstep_stream" "stream_2" {
  project_name = "${lightstep_project.project.project_name}"
  stream_name = "test_stream_2"
  query = "tag:\"error\"=\"false\""
  custom_data = {
    test_string = "Hello World"
    test_map = "This Cool"
  }
  depends_on = ["lightstep_project.project"]
}

resource "lightstep_project" "project" {
  project = "saladbar-terraform_test5"
}

resource "lightstep_dashboard" "test_dashboard" {
  project = "saladbar-terraform"
  name = "test_dashboard"
  search_attributes = [{
    name = "test_stream_2",
    query = "tag:\"error\"=\"false\""
  }, {
    name = "test_stream_1",
    query = "tag:\"error\"=\"true\""
  }]
}
