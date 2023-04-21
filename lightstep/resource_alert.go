package lightstep

import (
	"fmt"
	"strconv"

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

func getCompositeAlertFromUnifiedConditionResourceData(compositeAlertIn *client.CompositeAlert) (interface{}, error) {
	subAlerts := make([]map[string]interface{}, 0)
	for _, subAlertIn := range compositeAlertIn.Alerts {
		// TODO: refactor me
		thresholdEntries := map[string]interface{}{}
		if subAlertIn.Expression.Thresholds.Critical != nil {
			thresholdEntries["critical"] = strconv.FormatFloat(*subAlertIn.Expression.Thresholds.Critical, 'f', -1, 64)
		}

		if subAlertIn.Expression.Thresholds.Warning != nil {
			thresholdEntries["warning"] = strconv.FormatFloat(*subAlertIn.Expression.Thresholds.Warning, 'f', -1, 64)
		}

		queries, err := getQueriesFromUnifiedConditionResourceData(subAlertIn.Queries, subAlertIn.Name) // TODO: pass ID & name
		if err != nil {
			return nil, err
		}

		subAlerts = append(subAlerts, map[string]interface{}{
			"name":  subAlertIn.Name,
			"title": subAlertIn.Title,
			"expression": []map[string]interface{}{
				{
					"is_no_data": subAlertIn.Expression.IsNoData,
					"operand":    subAlertIn.Expression.Operand,
					"thresholds": []interface{}{
						thresholdEntries,
					},
				},
			},
			"query": queries,
		})
	}

	return []map[string][]map[string]interface{}{{
		"alert": subAlerts,
	}}, nil
}
