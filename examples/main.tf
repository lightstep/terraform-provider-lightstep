##############################################################
## Streams
##############################################################
resource "lightstep_stream" "custom_data" {
  project_name = var.project
  stream_name  = "custom_data_test0"
  query        = "operation IN (\"api/v1/charge\") AND \"customer_id\" NOT IN (\"test0\")"
  custom_data = [
    {
      # This name field is special and becomes the key
      "name" = "playbook"
      "url"  = "https://www.lightstep.com",
    },
  ]
}
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
  dashboard_name = "Customer Charges (Stream)"
  stream_ids     = [lightstep_stream.beemo.id, lightstep_stream.non_beemo.id]
}

resource "lightstep_metric_dashboard" "customer_charges" {
  project_name   = var.project
  dashboard_name = "Customer Charges (Metrics)"

  chart {
    name = "Requests by Project"
    rank = 1
    type = "timeseries"

    y_axis {
      min = 0.4
      max = 5.0
    }

    query {
      hidden              = false
      query_name          = "a"
      display             = "line"
      timeseries_operator = "rate"
      metric              = "requests"

      include_filters = [
        {
          key   = "service"
          value = "iOS"
        }
      ]

      group_by {
        aggregation_method = "max"
        keys               = ["project_name"]
      }
    }
  }
}

##############################################################
## Conditions
##############################################################

resource "lightstep_stream_condition" "beemo_errors" {
  project_name         = var.project
  condition_name       = "Charge errors for BEEMO"
  expression           = "err > 0.4"
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

resource "lightstep_alerting_rule" "beemo_ops_alerting_rule" {
  project_name    = var.project
  condition_id    = lightstep_stream_condition.beemo_ops.id
  destination_id  = lightstep_slack_destination.slack.id
  update_interval = "1h"
}

resource "lightstep_alerting_rule" "beemo_latency_alerting_rule" {
  project_name    = var.project
  condition_id    = lightstep_stream_condition.beemo_latency.id
  destination_id  = lightstep_pagerduty_destination.pd.id
  update_interval = "1h"
}
resource "lightstep_metric_condition" "beemo_requests" {
  project_name = var.project
  name         = "Beemo Low Requests"

  expression {
    evaluation_window   = "2m"
    evaluation_criteria = "on_average"
    is_multi            = true
    is_no_data          = true
    operand             = "below"
    thresholds {
      warning  = 10.0
      critical = 5.0
    }
  }

  metric_query {
    metric              = "requests"
    query_name          = "a"
    timeseries_operator = "delta"
    hidden              = false
    display             = "line"

    include_filters = [
      {
        key   = "kube_instance"
        value = "3"
      }
    ]

    group_by {
      aggregation_method = "max"
      keys               = ["key1", "key2"]
    }
  }

  alerting_rule {
    id              = lightstep_pagerduty_destination.pd.id
    update_interval = "1h"

    include_filters = [
      {
        key   = "kube_instance"
        value = "3"
      }
    ]
  }

  alerting_rule {
    id              = lightstep_webhook_destination.webhook.id
    update_interval = "1h"
    exclude_filters = [
      {
        key   = "kube_instance"
        value = "1"
      }
    ]
  }

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