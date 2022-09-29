package lightstep

import (
	"github.com/lightstep/terraform-provider-lightstep/client"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
)

func getUnifiedQuerySchema() map[string]*schema.Schema {
	sma := map[string]*schema.Schema{
		"hidden": {
			Type:     schema.TypeBool,
			Required: true,
		},
		"display": {
			Type:         schema.TypeString,
			Optional:     true,
			ValidateFunc: validation.StringInSlice([]string{"line", "area", "bar", "big_number", "heatmap"}, false),
		},
		"query_name": {
			Type:     schema.TypeString,
			Required: true,
		},
		"query_string": {
			Type:     schema.TypeString,
			Required: true,
		},
	}
	return sma
}

func getQueriesFromUnifiedDashboardResourceData(queriesIn []client.MetricQueryWithAttributes) []interface{} {
	var queries []interface{}
	for _, q := range queriesIn {
		qs := map[string]interface{}{
			"hidden":       q.Hidden,
			"display":      q.Display,
			"query_name":   q.Name,
			"query_string": q.TQLQuery,
		}
		queries = append(queries, qs)
	}
	return queries
}
