---
# generated by https://github.com/hashicorp/terraform-plugin-docs
page_title: "lightstep_metric_dashboard Resource - terraform-provider-lightstep"
subcategory: ""
description: |-

---

# lightstep_metric_dashboard (Resource)

Provides a [Lightstep Dashboard](https://api-docs.lightstep.com/reference/listmetricdashboardid). This can be used to create and manage Lightstep Dashboards.


## Example Usage

```hcl
resource "lightstep_dashboard" "customer_charges" {
  project_name   = var.project
  dashboard_name = "Customer Charges (Metrics)"

  chart {
    name = "Requests by Project"
    rank = 1
    type = "timeseries"

    query {
      hidden         = false
      query_name     = "a"
      display        = "line"
      query_string   = "metric requests | rate 10m | group_by [project_id], sum"
    }
  }

  chart {
    name = "Public API Latency"
    rank = "2"
    type = "timeseries"

    query {
      query_name     = "a"
      display        = "line"
      hidden         = false
      query_string   = "spans latency | rate 10m | group_by [customer_name], max"
    }
  }
}
```

<!-- schema generated by tfplugindocs -->
## Schema

### Required

- `dashboard_name` (String)
- `project_name` (String)

### Optional

- `chart` (Block Set) (see [below for nested schema](#nestedblock--chart))

### Read-Only

- `id` (String) The ID of this resource.
- `type` (String)

<a id="nestedblock--chart"></a>
### Nested Schema for `chart`

Required:

- `name` (String)
- `query` (Block List, Min: 1) (see [below for nested schema](#nestedblock--chart--query))
- `rank` (Number)
- `type` (String)

Optional:

- `y_axis` (Block List, Max: 1, Deprecated) (see [below for nested schema](#nestedblock--chart--y_axis))

Read-Only:

- `id` (String) The ID of this resource.

<a id="nestedblock--chart--query"></a>
### Nested Schema for `chart.query`

Required:

- `hidden` (Boolean)
- `query_name` (String)
- `query_string` (String)

Optional:

- `display` (String)


<a id="nestedblock--chart--y_axis"></a>
### Nested Schema for `chart.y_axis`

Required:

- `max` (Number)
- `min` (Number)