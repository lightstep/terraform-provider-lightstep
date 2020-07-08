provider "lightstep" {
  organization = "LightStep"
}

#############################################################
# Streams
#############################################################
//resource "lightstep_stream" "stream_1" {
//  project_name = "${lightstep_project.demo_project.project_name}"
//  stream_name = "BEEMO Charges"
//  query = "tag:\"customer_id\"=\"BEEMO\" operation:\"api/v1/charge\""
//}
//
//resource "lightstep_stream" "stream_2" {
//  project_name = "${lightstep_project.demo_project.project_name}"
//  stream_name = "BEEMO Reserve Assets"
//  query = "tag:\"customer_id\"=\"BEEMO\" operation:\"api/v1/reserve-asset\""
//  custom_data = {
//    legacy = "false"
//  }
//}
//
//resource "lightstep_stream" "stream_3" {
//  project_name = "${lightstep_project.demo_project.project_name}"
//  stream_name = "BEEMO Webapp Load"
//  query = "tag:\"customer_id\"=\"BEEMO\" operation:\"dom-load\""
//}
//
//resource "lightstep_stream" "stream_4" {
//  project_name = "${lightstep_project.demo_project.project_name}"
//  stream_name = "BEEMO Start Reservation Flow"
//  query = "tag:\"customer_id\"=\"BEEMO\" operation:\"start-reservation-flow\""
//}
//
//resource "lightstep_stream" "stream_5" {
//  project_name = "${lightstep_project.demo_project.project_name}"
//  stream_name = "ACME Charges"
//  query = "tag:\"customer_id\"=\"ACME\" operation:\"api/v1/charge\""
//}
//
//resource "lightstep_stream" "stream_6" {
//  project_name = "${lightstep_project.demo_project.project_name}"
//  stream_name = "ACME Reserve Assets"
//  query = "tag:\"customer_id\"=\"ACME\" operation:\"api/v1/reserve-asset\""
//  custom_data = {
//    legacy = "false"
//  }
//}
//
//resource "lightstep_stream" "stream_7" {
//  project_name = "${lightstep_project.demo_project.project_name}"
//  stream_name = "ACME Webapp Load"
//  query = "tag:\"customer_id\"=\"ACME\" operation:\"dom-load\""
//}
//
//resource "lightstep_stream" "stream_8" {
//  project_name = "${lightstep_project.demo_project.project_name}"
//  stream_name = "ACME Start Reservation Flow"
//  query = "tag:\"customer_id\"=\"ACME\" operation:\"start-reservation-flow\""
//}
//
//#############################################################
//# Dashboards
//#############################################################
//resource "lightstep_dashboard" "dashboard" {
//  project_name = "${lightstep_project.demo_project.project_name}"
//  dashboard_name = "BEEMO -- #TERRAFORM FTW"
//  streams = [
//    {
//      stream_name = "${lightstep_stream.stream_1.stream_name}"
//      query = "${lightstep_stream.stream_1.query}"
//    },
//    {
//      stream_name = "${lightstep_stream.stream_2.stream_name}"
//      query = "${lightstep_stream.stream_2.query}"
//    },
//    {
//      stream_name = "${lightstep_stream.stream_3.stream_name}"
//      query = "${lightstep_stream.stream_3.query}"
//    },
//    {
//      stream_name = "${lightstep_stream.stream_4.stream_name}"
//      query = "${lightstep_stream.stream_4.query}"
//    },
//  ]
//}
//
//resource "lightstep_dashboard" "dashboard_acme" {
//  project_name = "${lightstep_project.demo_project.project_name}"
//  dashboard_name = "ACME -- #TERRAFORM FTW"
//  streams = [
//    {
//      stream_name = "${lightstep_stream.stream_5.stream_name}"
//      query = "${lightstep_stream.stream_5.query}"
//    },
//    {
//      stream_name = "${lightstep_stream.stream_6.stream_name}"
//      query = "${lightstep_stream.stream_6.query}"
//    },
//    {
//      stream_name = "${lightstep_stream.stream_7.stream_name}"
//      query = "${lightstep_stream.stream_7.query}"
//    },
//    {
//      stream_name = "${lightstep_stream.stream_8.stream_name}"
//      query = "${lightstep_stream.stream_8.query}"
//    },
//  ]
//}
