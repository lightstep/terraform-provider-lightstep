package lightstep

import (
	"context"
	"fmt"
	"regexp"
	"strings"

	"github.com/lightstep/terraform-provider-lightstep/client"
)

// Checks if the prior dashboard attributes have legacy queries that
// map to the UQL queries that the API call returns.
//
// If so, we can ignore the "false diff" a Terraform plan would otherwise show
// by using the legacy resource data instread of the UQL resource data.
//
// By design, this checks *only* the query equivalence, not other attributes.
func dashboardHasEquivalentLegacyQueries(
	ctx context.Context,
	c *client.Client,
	projectName string,
	priorCharts []client.UnifiedChart,
	updatedCharts []client.UnifiedChart,
) (bool, error) {
	// This code is only applicable if there are legacy charts
	allTQL := true
	for _, chart := range priorCharts {
		if !hasOnlyTQLQueries(chart.MetricQueries) {
			allTQL = false
			break
		}
	}
	if allTQL {
		return false, nil
	}

	if len(priorCharts) != len(updatedCharts) {
		// if there's a different number of charts, the queries can't possibly be equivalent
		return false, nil
	}

	// Loop through each chart and compare...
	for _, priorChart := range priorCharts {
		var updatedChart *client.UnifiedChart
		for _, chart := range updatedCharts {
			// use rank as a unique identifier as IDs may not yet be known
			if priorChart.Rank == chart.Rank {
				updatedChart = &chart
				break
			}
		}
		if updatedChart == nil {
			// this chart doesnt exist anymore, do an update
			return false, nil
		}

		// Check the converted query
		equivalent, err := compareUpdatedLegacyQueries(
			ctx, c, projectName,
			priorChart.MetricQueries,
			updatedChart.MetricQueries,
		)
		if err != nil {
			return false, err
		}
		if !equivalent {
			return false, nil
		}
	}
	return true, nil
}

// See dashboardHasEquivalentLegacyQueries: this is the metric condition
// version
//
// By design, this checks *only* the query equivalence, not other attributes.
func metricConditionHasEquivalentLegacyQueries(
	ctx context.Context,
	c *client.Client,
	projectName string,
	priorAttrs *client.UnifiedConditionAttributes,
	updatedAttrs *client.UnifiedConditionAttributes,
) (bool, error) {
	// This code is only applicable if there are legacy charts. If it's
	// all TQL, return false to default to the normal code path
	if hasOnlyTQLQueries(priorAttrs.Queries) {
		return false, nil
	}

	// Compare the queries
	equivalent, err := compareUpdatedLegacyQueries(
		ctx, c, projectName,
		priorAttrs.Queries,
		updatedAttrs.Queries,
	)
	if err != nil {
		return false, err
	}
	return equivalent, nil
}

func hasOnlyTQLQueries(queries []client.MetricQueryWithAttributes) bool {
	for _, query := range queries {
		if query.Type != "tql" {
			return false
		}
	}
	return true
}

// Check that the prior and updated set of queries are equivalent by
// checking if the the UQL translations are the same.
func compareUpdatedLegacyQueries(
	ctx context.Context,
	c *client.Client,
	projectName string,
	priorQueries []client.MetricQueryWithAttributes,
	updatedQueries []client.MetricQueryWithAttributes,
) (bool, error) {
	// Step 0: if these are both TQL queries, the strings should match
	// exactly. (Note that a single query-set with a mix of TQL and
	// legacy has undefined behavior so we do not check for that case.)
	if len(priorQueries) == len(updatedQueries) &&
		hasOnlyTQLQueries(priorQueries) && hasOnlyTQLQueries(updatedQueries) {
		for i, pq := range priorQueries {
			if pq.TQLQuery != updatedQueries[i].TQLQuery {
				return false, nil
			}
		}
	}

	// Step 1: call the SaaS to translate the legacy queries to UQL
	type Response struct {
		Data struct {
			Queries []struct {
				QueryName   string `json:"query-name"`
				QueryString string `json:"tql-query"`
			} `json:"queries"`
		} `json:"data"`
	}

	resp := Response{}
	req := map[string]interface{}{
		"data": map[string]interface{}{
			"queries": priorQueries,
		},
	}
	err := c.CallAPI(context.Background(), "POST", fmt.Sprintf("projects/%v/query_translation", projectName), req, &resp)
	if err != nil {
		return false, err
	}

	priorUQL := make(map[string]string)
	for _, q := range resp.Data.Queries {
		priorUQL[simplifyQueryName(q.QueryName)] = q.QueryString
	}

	// Step 2: map the updated quries for comparison
	updatedUQL := make(map[string]string)
	for _, q := range updatedQueries {
		if q.Type != "tql" {
			return false, nil
		}
		updatedUQL[simplifyQueryName(q.Name)] = q.TQLQuery
	}

	// Step 3: compare the queries are equivalent
	if len(priorUQL) != len(updatedUQL) {
		return false, nil
	}
	for key, prior := range priorUQL {
		updated, ok := updatedUQL[key]
		if !ok {
			return false, nil
		}
		prior = strings.TrimSpace(prior)
		updated = strings.TrimSpace(updated)
		if prior != updated {
			return false, nil
		}
	}
	return true, nil
}

var simplifyQueryNameRE = regexp.MustCompile(`[()\s]`)

// Simplifies a query name as query names can be formulas and the
// whitespace and use of parens is not always consistent, so ignore
// those differences in the naming.
func simplifyQueryName(s string) string {
	return simplifyQueryNameRE.ReplaceAllString(s, "")
}
