provider "lightstep" {
  organization = "LightStep"
}

resource "lightstep_project" "project" {
  project_name = "saladbar-terraform_test28"
}

# Streams

resource "lightstep_stream" "stream_1" {
  project_name = "${lightstep_project.project.project_name}"
  stream_name = "test_stream_1"
  query = "tag:\"error\"=\"true\""
}

resource "lightstep_stream" "stream_2" {
  project_name = "${lightstep_project.project.project_name}"
  stream_name = "test_stream_ishmeet"
  query = "tag:\"error\"=\"false\""
  custom_data = {
    test_string = "Hello Ishmeet"
    test_map = "This Cool"
  }
}

resource "lightstep_stream" "stream_3" {
  project_name = "${lightstep_project.project.project_name}"
  stream_name = "test_stream_3"
  query = "tag:\"release_tag\"=\"373bef115c81f552\""
}

resource "lightstep_stream" "stream_4" {
  project_name = "${lightstep_project.project.project_name}"
  stream_name = "test_stream_4"
  query = "tag:\"release_tag\"=\"ef11373b5c5281f5\""
}

resource "lightstep_stream" "stream_5" {
  project_name = "${lightstep_project.project.project_name}"
  stream_name = "test_stream_5"
  query = "tag:\"release_tag\"=\"ishmeet\""
}

# Dashboard

resource "lightstep_dashboard" "dashboard" {
  project_name = "${lightstep_project.project.project_name}"
  dashboard_name = "test_dashboard"
  streams = [
    {
      stream_name = "${lightstep_stream.stream_1.stream_name}"
      query = "${lightstep_stream.stream_1.query}"
    },
    {
      stream_name = "${lightstep_stream.stream_2.stream_name}"
      query = "${lightstep_stream.stream_2.query}"
    },
    {
      stream_name = "${lightstep_stream.stream_3.stream_name}"
      query = "${lightstep_stream.stream_3.query}"
    },
    {
      stream_name = "${lightstep_stream.stream_4.stream_name}"
      query = "${lightstep_stream.stream_4.query}"
    },
    {
      stream_name = "${lightstep_stream.stream_5.stream_name}"
      query = "${lightstep_stream.stream_5.query}"
    },
  ]
}
