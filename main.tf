provider "lightstep" {
  organization = "LightStep"
}

#############################################################
# Streams
#############################################################

resource "lightstep_stream" "stream_2" {
  project_name = "dev-paigebernier"
  stream_name = "BEEMO Charges new new"
  query = "operation IN (\"api/v1/charge\") AND \"customer_id\" IN (\"BEEMO\")"
}
