provider "lightstep" {
  organization = "LightStep"
}

#############################################################
# Projects
#############################################################
resource "lightstep_project" "demo_project" {
  project_name = "nksingh-demo"
}

#############################################################
# Streams
#############################################################
resource "lightstep_stream" "stream_1" {
  project_name = "${lightstep_project.demo_project.project_name}"
  stream_name = "BEEMO Charges"
  query = "tag:\"customer_id\"=\"BEEMO\" operation:\"api/v1/charge\""
}

resource "lightstep_stream" "stream_2" {
  project_name = "${lightstep_project.demo_project.project_name}"
  stream_name = "BEEMO Reserve Assets"
  query = "tag:\"customer_id\"=\"BEEMO\" operation:\"api/v1/reserve-asset\""
  custom_data = {
    legacy = "false"
  }
}

resource "lightstep_stream" "stream_3" {
  project_name = "${lightstep_project.demo_project.project_name}"
  stream_name = "BEEMO Webapp Load"
  query = "tag:\"customer_id\"=\"BEEMO\" operation:\"dom-load\""
}

resource "lightstep_stream" "stream_4" {
  project_name = "${lightstep_project.demo_project.project_name}"
  stream_name = "BEEMO Start Reservation Flow"
  query = "tag:\"customer_id\"=\"BEEMO\" operation:\"start-reservation-flow\""
}

#############################################################
# Dashboards
#############################################################
resource "lightstep_dashboard" "dashboard" {
  project_name = "${lightstep_project.demo_project.project_name}"
  dashboard_name = "BEEMO -- #TERRAFORM FTW"
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
  ]
}
