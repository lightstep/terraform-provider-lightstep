package exporter

import (
	"context"
	"fmt"
	"io"
	"log"
	"os"
	"strings"
	"text/template"

	"github.com/lightstep/terraform-provider-lightstep/client"
)

const metricDashboardTemplate = `
resource "lightstep_metric_dashboard" "exported_dashboard" {
  project_name = var.project
  dashboard_name = "{{.Attributes.Name}}"
  dashboard_description = {{escapeHeredocString .Attributes.Description}}
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
         query         = "{{escapeHCLString .SpansQuery.Query}}"
         operator      = "{{.SpansQuery.Operator}}"
         group_by_keys = [{{range .SpansQuery.GroupByKeys}}"{{.}}",{{end}}]{{if eq .SpansQuery.Operator "latency"}}
         latency_percentiles = [{{range .SpansQuery.LatencyPercentiles}}{{.}},{{end}}]{{end}}
      }
{{end}}{{if .TQLQuery}}
      tql                 = {{escapeHeredocString .TQLQuery}}
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

const unifiedDashboardTemplate = `
resource "lightstep_dashboard" "exported_dashboard" {
  project_name = var.project
  dashboard_name = "{{.Attributes.Name}}"
  dashboard_description = {{escapeHeredocString .Attributes.Description}}
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
      query_string        = {{escapeHeredocString .TQLQuery}}
{{- if .DependencyMapOptions}}
      dependency_map_options {
        scope    = "{{.DependencyMapOptions.Scope}}"
        map_type = "{{.DependencyMapOptions.MapType}}"
      }
{{- end}}
    }
{{end}}
  }
{{end}}
}
`

func escapeHCLString(input string) string {
	// Escape "\" first so other the other escape codes don't get escaped
	input = strings.Replace(input, "\\", "\\\\", -1)

	input = strings.Replace(input, "\"", "\\\"", -1)
	input = strings.Replace(input, "\n", "\\\t", -1)
	input = strings.Replace(input, "\r", "\\\t", -1)
	input = strings.Replace(input, "\t", "\\\t", -1)
	return input
}

func escapeHeredocString(input string) string {
	// Use "heredoc" syntax if the query contains any newlines or other characters that'd
	// need to be escaped and make the single line representation less convenient to work with.
	if strings.Contains(input, "\"") ||
		strings.Contains(input, "\n") ||
		strings.Contains(input, "\r") ||
		strings.Contains(input, "\t") ||
		strings.Contains(input, "\\") {
		return "<<EOT\n" + input + "\nEOT"
	} else {
		// No need to escape input since the other branch should be hit for any strings requiring
		// escaping.
		return `"` + input + `"`
	}
}

// dashboardUsesLegacyQuery returns true if any chart in the dashboard
// uses the legacy query format.
func dashboardUsesLegacyQuery(d *client.UnifiedDashboard) bool {
	for _, chart := range d.Attributes.Charts {
		for _, q := range chart.MetricQueries {
			// Assume if a chart is defined but has query string defined, it uses a legacy query.  This
			// isn't strictly correct if the chart has *no* query but a chart with no query is not
			// meaningful to begin with.
			if len(q.TQLQuery) == 0 {
				return true
			}
		}
	}
	return false
}

func exportToHCL(wr io.Writer, d *client.UnifiedDashboard) error {
	t := template.New("").Funcs(template.FuncMap{
		"escapeHCLString":     escapeHCLString,
		"escapeHeredocString": escapeHeredocString,
	})

	usesLegacyQuery := dashboardUsesLegacyQuery(d)

	// Use the legacy format if any chart uses a legacy query
	var hclTemplate string
	if usesLegacyQuery {
		hclTemplate = metricDashboardTemplate
	} else {
		hclTemplate = unifiedDashboardTemplate
	}
	t, err := t.Parse(hclTemplate)
	if err != nil {
		return fmt.Errorf("dashboard parsing error: %v", err)
	}

	err = t.Execute(wr, d)
	if err != nil {
		log.Fatalf("Could not generate template: %v", err)
	}
	return nil
}

func Run(args ...string) error {
	if len(os.Getenv("LIGHTSTEP_API_KEY")) == 0 {
		log.Fatalf("error: LIGHTSTEP_API_KEY env variable must be set")
	}

	if len(os.Getenv("LIGHTSTEP_ORG")) == 0 {
		log.Fatalf("error: LIGHTSTEP_ORG env variable must be set")
	}

	// default to public API environment
	lightstepUrl := "https://api.lightstep.com"
	if len(os.Getenv("LIGHTSTEP_API_URL")) > 0 {
		lightstepUrl = os.Getenv("LIGHTSTEP_API_URL")
	} else if len(os.Getenv("LIGHTSTEP_API_BASE_URL")) > 0 {
		lightstepUrl = os.Getenv("LIGHTSTEP_API_BASE_URL")
	}

	if len(args) < 4 {
		log.Fatalf("usage: %s export [resource-type] [project-name] [resource-id]", args[0])
	}

	if args[2] != "dashboard" && args[2] != "lightstep_dashboard" {
		log.Fatalf("error: only dashboard resources are supported at this time")
	}

	c := client.NewClient(os.Getenv("LIGHTSTEP_API_KEY"), os.Getenv("LIGHTSTEP_ORG"), lightstepUrl)
	d, err := c.GetUnifiedDashboard(context.Background(), args[3], args[4])
	if err != nil {
		log.Fatalf("error: could not get dashboard: %v", err)
	}

	err = exportToHCL(os.Stdout, d)
	if err != nil {
		log.Fatalf("Could not export to HCL: %v", err)
	}

	return nil
}
