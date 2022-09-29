package exporter

import (
	"context"
	"log"
	"os"
	"strings"
	"text/template"

	"github.com/lightstep/terraform-provider-lightstep/client"
)

const dashboardTemplate = `
resource "lightstep_metric_dashboard" "exported_dashboard" {
  project_name = var.project
  dashboard_name = "{{.Attributes.Name}}"
{{range .Attributes.Charts}}
  chart {
    name = "{{.Title}}"
    rank = "{{.Rank}}"
    type = "{{.ChartType}}"
{{range .MetricQueries}}
    query {
      query_name          = "{{.Name}}"
      display             = "{{.Display}}"
      hidden              = {{.Hidden}}
{{if (and .SpansQuery .SpansQuery.Query) }}
      spans {
         query         = "{{escapeSpanQuery .SpansQuery.Query}}"
         operator      = "{{.SpansQuery.Operator}}"
         group_by_keys = [{{range .SpansQuery.GroupByKeys}}"{{.}}",{{end}}]{{if eq .SpansQuery.Operator "latency"}}
         latency_percentiles = [{{range .SpansQuery.LatencyPercentiles}}{{.}},{{end}}]{{end}}
      }
{{end}}{{if .TQLQuery}}
      tql                 = "{{.TQLQuery}}"
{{end}}{{if .Query.Metric}}
      metric              = "{{.Query.Metric}}"
      timeseries_operator = "{{.Query.TimeseriesOperator}}"
{{if .Query.Filters}}
      include_filters = [{{range .Query.Filters}}
        {
          key   = "{{.Key}}"
          value = "{{.Value}}"
        },{{end}}
      ]
{{end}}
{{if .Query.GroupBy}}
      group_by {
        aggregation_method = "{{.Query.GroupBy.Aggregation}}"
        keys = [{{range .Query.GroupBy.LabelKeys}}"{{.}}",{{end}}]
      }{{end}}
{{end}}
    }
{{end}}
  }
{{end}}
}
`

func escapeSpanQuery(input string) string {
	return strings.Replace(input, "\"", "\\\"", -1)
}

func Run(args ...string) error {
	if len(os.Getenv("LIGHTSTEP_API_KEY")) == 0 {
		log.Fatalf("error: LIGHTSTEP_API_KEY env variable must be set")
	}

	if len(os.Getenv("LIGHTSTEP_ORG")) == 0 {
		log.Fatalf("error: LIGHTSTEP_ORG env variable must be set")
	}

	// default to public API environment
	lightstepEnv := "public"
	if len(os.Getenv("LIGHTSTEP_ENV")) > 0 {
		lightstepEnv = os.Getenv("LIGHTSTEP_ENV")
	}

	if len(args) < 4 {
		log.Fatalf("usage: %s export [resource-type] [project-name] [resource-id]", args[0])
	}

	if args[2] != "dashboard" {
		log.Fatalf("error: only dashboard resources are supported at this time")
	}

	c := client.NewClient(os.Getenv("LIGHTSTEP_API_KEY"), os.Getenv("LIGHTSTEP_ORG"), lightstepEnv)
	d, err := c.GetUnifiedDashboard(context.Background(), args[3], args[4])

	if err != nil {
		log.Fatalf("error: could not get dashboard: %v", err)
	}

	t := template.New("").Funcs(template.FuncMap{
		"escapeSpanQuery": escapeSpanQuery,
	})

	t, err = t.Parse(dashboardTemplate)
	if err != nil {
		log.Fatal("Dashboard parsing error: ", err)
	}

	err = t.Execute(os.Stdout, d)
	if err != nil {
		log.Fatalf("Could not generate template: %v", err)
	}

	return nil
}
