package exporter

import (
	"bytes"
	"fmt"
	"strings"
	"testing"

	"github.com/lightstep/terraform-provider-lightstep/client"
)

func TestExportToHCL(t *testing.T) {
	export := func(d *client.UnifiedDashboard) (string, error) {
		var buf bytes.Buffer
		err := exportToHCL(&buf, d)
		return buf.String(), err
	}

	testCases := []struct {
		QueryString string
		Expected    string
	}{
		{
			QueryString: "",
			Expected:    "",
		},
		{
			QueryString: "metric requests | rate 10m",
			Expected:    `query_string        = "metric requests | rate 10m"`,
		},
		{
			QueryString: `metric requests | rate | filter service = "apache" | group_by [], sum`,
			Expected: `query_string        = <<EOT
metric requests | rate | filter service = "apache" | group_by [], sum
EOT`,
		},
		{
			QueryString: `metric requests | 
	rate | 
	group_by [], sum
`,
			Expected: `query_string        = <<EOT
metric requests | 
	rate | 
	group_by [], sum

EOT`,
		},
		{
			QueryString: "metric w\t|rate 10m",
			Expected: `query_string        = <<EOT
metric w	|rate 10m
EOT`,
		},
	}

	for index, testCase := range testCases {
		t.Run(fmt.Sprintf("test case %v", index), func(t *testing.T) {

			s, err := export(&client.UnifiedDashboard{
				Attributes: client.UnifiedDashboardAttributes{
					Name: "Test dashboard",
					Charts: []client.UnifiedChart{
						{
							MetricQueries: []client.MetricQueryWithAttributes{
								{
									Name:     "my_query",
									Display:  "line",
									Hidden:   false,
									TQLQuery: testCase.QueryString,
								},
							},
						},
					},
				},
			})
			if err != nil {
				t.Errorf("unexpected error: %v", err)
			}

			if !strings.Contains(s, testCase.Expected) {
				fmt.Printf("Expected:\n%v\n", testCase.Expected)
				fmt.Printf("HCL:\n%v\n", s)
				t.Errorf("resulting HCL does not contain the expected substring")
			}
		})
	}
}
