provider "lightstep" {
  environment     = "staging"
  api_key_env_var = "LIGHTSTEP_STAGING_API_KEY"
  organization    = "LightStep"
}

variable "project" {
  type    = string
  default = "YOUR PROJECT HERE"
}

##############################################################
## Streams
##############################################################
resource "lightstep_stream" "non_beemo" {
  project_name = var.project
  stream_name  = "Non-BEEMO charges"
  query        = "operation IN (\"api/v1/charge\") AND \"customer_id\" NOT IN (\"BEEMO\")"
}

resource "lightstep_stream" "beemo" {
  project_name = var.project
  stream_name  = "BEEMO charges"
  query        = "operation IN (\"api/v1/charge\") AND \"customer_id\" IN (\"BEEMO\")"
}

##############################################################
## Dashboards
##############################################################

resource "lightstep_stream_dashboard" "customer_charges" {
  project_name   = var.project
  dashboard_name = "Customer Charges"
  stream_ids     = [lightstep_stream.beemo.id, lightstep_stream.non_beemo.id]
}

resource "lightstep_stream_dashboard" "customer_charges" {
  project_name   = var.project
  dashboard_name = "Customer Charges"
  stream_ids     = [lightstep_stream.beemo.id, lightstep_stream.non_beemo.id]
}

##############################################################
## Conditions
##############################################################

resource "lightstep_stream_condition" "beemo_errors" {
  project_name         = var.project
  condition_name       = "Charge errors for BEEMO"
  expression           = "err > .4"
  evaluation_window_ms = 300000
  stream_id            = lightstep_stream.beemo.id
}

resource "lightstep_stream_condition" "beemo_latency" {
  project_name         = var.project
  condition_name       = "High Latency for Charge to BEEMO"
  expression           = "lat(95) > 5s"
  evaluation_window_ms = 300000
  stream_id            = lightstep_stream.beemo.id
}

resource "lightstep_stream_condition" "beemo_ops" {
  project_name         = var.project
  condition_name       = "Abnormally low ops for BEEMO charge"
  expression           = "ops < 100"
  evaluation_window_ms = 1200000 # 20 minutes
  stream_id            = lightstep_stream.beemo.id
}

resource "lightstep_metric_condition" "test" {
  project_name   = var.project
  condition_name = "Too many requests"

  evaluation_window   = 120000000
  evaluation_criteria = "on_average"

  display = "line"

  is_multi   = true
  is_no_data = true

  thresholds = {
    operand  = "above"
    critical = 10
    warning  = 5
  }

  query {
    metric_name         = "requests"
    query_name          = "a"
    type                = "single"
    timeseries_operator = "rate"
    hidden              = false

    include_filters = [{
      key   = "project_name"
      value = "catlab"
    }]

    exclude_filters = [{
      key   = "service"
      value = "android"
    }]

    group_by {
      aggregation = "avg"
      keys        = ["method"]
    }
  }
>>>>>>> metricConditions can be created, read, deleted (without alertingRules)
}

##############################################################
## Destinations
##############################################################

resource "lightstep_webhook_destination" "webhook" {
  project_name     = var.project
  destination_name = "my svc"
  url              = "https://www.downforeveryoneorjustme.com"

  custom_headers = {
    "Cache-Control"   = "max-age=0"
    "Referrer-Policy" = "no-referrer"
  }

}

resource "lightstep_pagerduty_destination" "pd" {
  project_name     = var.project
  destination_name = "My Destination"
  integration_key  = "eec7e430f6gd489b8e91ebcae17a3f42"
}

resource "lightstep_slack_destination" "slack" {
  project_name = var.project
  channel      = "#urgent-care"
}
