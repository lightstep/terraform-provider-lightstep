package lightstep

import (
	"fmt"

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

func getQueriesFromUnifiedDashboardResourceData(
	queriesIn []client.MetricQueryWithAttributes,
	dashboardID string,
	chartID string,
) ([]interface{}, error) {

	var queries []interface{}
	for _, q := range queriesIn {
		// Check if a query is using the legacy representation.  In that case,
		// it can't be represented in the "query_string" field and the call should
		// fail. In the future, we should attempt to automatically convert legacy -> query
		// string format. At the moment, there's no public API to implement this, so
		// at least provide a clarifying error message.
		if q.Type != "tql" {
			return nil, fmt.Errorf(
				"cannot convert query from chart %v in dashboard %v\n\n"+
					"Query is of type '%v' but must be of type 'tql' for use with the resource\n"+
					"type lightstep_dashboard.\n"+
					"\n"+
					"Try using the lightstep_metrics_dashboard resource type for this dashboard\n"+
					"or convert all queries in the dashboard to query string format. ",
				chartID,
				dashboardID,
				q.Type,
			)
		}

		qs := map[string]interface{}{
			"hidden":       q.Hidden,
			"display":      q.Display,
			"query_name":   q.Name,
			"query_string": q.TQLQuery,
		}
		queries = append(queries, qs)
	}
	return queries, nil
}
