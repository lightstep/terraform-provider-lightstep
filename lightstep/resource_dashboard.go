package lightstep

import (
	"fmt"

	"github.com/lightstep/terraform-provider-lightstep/client"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
)

func getUnifiedQuerySchemaMap() map[string]*schema.Schema {
	sma := map[string]*schema.Schema{
		"hidden": {
			Type:     schema.TypeBool,
			Required: true,
		},
		"display": {
			Type:     schema.TypeString,
			Optional: true,
			ValidateFunc: validation.StringInSlice([]string{
				"line",
				"area",
				"bar",
				"big_number",
				"heatmap",
				"dependency_map",
				"big_number_v2",
				"scatter_plot",
				"ordered_list",
				"table",
			}, false),
		},
		// See https://github.com/hashicorp/terraform-plugin-sdk/issues/155
		// Using a TypeSet of size 1 as a way to allow nested properties
		"display_type_options": {
			Type:     schema.TypeSet,
			Optional: true,
			MaxItems: 1,
			Elem: &schema.Resource{
				// This is the superset of all possible fields for all display types
				Schema: map[string]*schema.Schema{
					"sort_by": {
						Type:     schema.TypeString,
						Optional: true,
					},
					"sort_direction": {
						Type:     schema.TypeString,
						Optional: true,
					},
				},
			},
		},
		"query_name": {
			Type:     schema.TypeString,
			Required: true,
		},
		"query_string": {
			Type:     schema.TypeString,
			Required: true,
		},
		"hidden_queries": {
			Type: schema.TypeMap,
			Elem: &schema.Schema{
				Type: schema.TypeString,
			},
			Optional: true,
			Description: "An optional map of sub-query names in the query_string to a boolean string to hide/show that query. " +
				"If specified, the map must have an entry for all named sub-queries in the query_string. A value " +
				"of \"true\" indicates the query should be hidden. " +
				"Example: `hidden_queries = {  \"a\" = \"true\",  \"b\" = \"false\" }`.",
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
			"hidden":                 q.Hidden,
			"display":                q.Display,
			"display_type_options":   displayTypeOptionsFromResourceData(q.DisplayTypeOptions),
			"query_name":             q.Name,
			"query_string":           q.TQLQuery,
			"dependency_map_options": getDependencyMapOptions(q.DependencyMapOptions),
		}
		if len(q.HiddenQueries) > 0 {
			// Note due to Terraform's issues with TypeMap having TypeBool elements, we
			// need to use boolean strings
			hq := make(map[string]interface{}, len(q.HiddenQueries)+1)
			for k, v := range q.HiddenQueries {
				// Don't include the top-level query in the TF resource data
				if k == q.Name {
					continue
				}
				hq[k] = fmt.Sprintf("%t", v)
			}
			qs["hidden_queries"] = hq
		}

		queries = append(queries, qs)
	}
	return queries, nil
}

func displayTypeOptionsFromResourceData(opts map[string]interface{}) *schema.Set {
	// "display_type_options" is a set that always has at most one element, so
	// the hash function is trivial
	f := func(i interface{}) int {
		return 1
	}
	return schema.NewSet(f, []interface{}{opts})
}

func getDependencyMapOptions(options *client.DependencyMapOptions) []interface{} {
	if options == nil {
		return nil
	}

	return []interface{}{
		map[string]interface{}{
			"map_type": options.MapType,
			"scope":    options.Scope,
		},
	}
}
