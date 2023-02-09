package lightstep

import (
	"fmt"
	"github.com/lightstep/terraform-provider-lightstep/client"
)

// most of the code that backs the lightstep_alert resource is shared with the legacy resource_metric_condition,
// and can be found in resource_metric_condition.go

func getQueriesFromUnifiedConditionResourceData(
	queriesIn []client.MetricQueryWithAttributes,
	conditionID string,
) ([]interface{}, error) {
	var queries []interface{}
	for _, q := range queriesIn {
		if q.Type != "tql" {
			return nil, fmt.Errorf(
				"cannot convert query from condition %v\n\n"+
					"Query is of type '%v' but must be of type 'tql' for use with the resource\n"+
					"type lightstep_alert.\n"+
					"\n"+
					"Try using the lightstep_metrics_condition resource type for this condition\n"+
					"or convert all queries in the condition to query string format. ",
				conditionID,
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
