package client

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func Test_getUnifiedDashboardURL(t *testing.T) {
	testCases := []struct {
		projID   [2]string
		query    map[string]string
		expected string
	}{
		{
			[2]string{"my_project", "123"}, nil,
			"projects/my_project/metric_dashboards/123",
		},
		{
			[2]string{"ProductionEnvironment", "fLx72349023"}, nil,
			"projects/ProductionEnvironment/metric_dashboards/fLx72349023",
		},
		{
			[2]string{"my_project", "123"}, map[string]string{"query_format": "query_string"},
			"projects/my_project/metric_dashboards/123?query_format=query_string",
		},
		{
			[2]string{"my_project", "123"}, map[string]string{"a": "1", "b": "2", "c": "3"},
			"projects/my_project/metric_dashboards/123?a=1&b=2&c=3",
		},
	}
	for _, c := range testCases {
		result := getUnifiedDashboardURL(c.projID[0], c.projID[1], c.query)
		require.Equal(t, c.expected, result)
	}
}
