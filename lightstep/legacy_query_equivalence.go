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
// If so, we can ignore the "false diff" a Terraform plan would otherwise show.
// This provides backwards compatibility for the legacy format (at least until
// the query is modified, at which point it needs to be converted to UQL)
func hasLegacyQueriesEquivalentToTQL(
	c *client.Client,
	projectName string,
	priorAttrs *client.UnifiedDashboardAttributes,
	updatedAttrs *client.UnifiedDashboardAttributes,
) (bool, error) {
	// This code is only applicalbe if there legacy charts
	if !hasLegacyQueries(priorAttrs) {
		return false, nil
	}

	// Loop through each chart and compare the queries...
	for index, priorChart := range priorAttrs.Charts {
		// Assumes the order of charts is _not_ going to change in the internal
		// data structure.  Note that we can't do an Chart.ID look up since the
		// prior structure doesn't necessarily have an ID at this point.
		updateChart := updatedAttrs.Charts[index]
		ok, err := compareUpdatedLegacyChart(c, projectName, priorChart, updateChart)
		if err != nil {
			return false, err
		}
		if !ok {
			return false, nil
		}
	}
	return true, nil

}

// Returns true if any of the queries on the dashboard are defined
// in legacy format.
func hasLegacyQueries(attrs *client.UnifiedDashboardAttributes) bool {
	for _, chart := range attrs.Charts {
		for _, query := range chart.MetricQueries {
			if query.Type != "tql" {
				return true
			}
		}
	}
	return false
}

func compareUpdatedLegacyChart(
	c *client.Client,
	projectName string,
	priorChart client.UnifiedChart,
	updatedChart client.UnifiedChart,
) (bool, error) {

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
			"queries": priorChart.MetricQueries,
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
	for _, q := range updatedChart.MetricQueries {
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
