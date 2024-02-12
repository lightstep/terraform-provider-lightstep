package exporter

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/lightstep/terraform-provider-lightstep/client"
	"github.com/stretchr/testify/assert"
)

func TestExportToHCL(t *testing.T) {
	export := func(d *client.UnifiedDashboard) (string, error) {
		var buf bytes.Buffer
		err := exportToHCL(&buf, d)
		return buf.String(), err
	}

	testCases := []struct {
		QueryString          string
		DependencyMapOptions *client.DependencyMapOptions
		Expected             string
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
		{
			QueryString: "spans_sample service = apache | assemble",
			DependencyMapOptions: &client.DependencyMapOptions{
				Scope:   "all",
				MapType: "service",
			},
			Expected: `query_string        = "spans_sample service = apache | assemble"
      dependency_map_options {
        scope    = "all"
        map_type = "service"
      }`,
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
									Name:                 "my_query",
									Display:              "line",
									Hidden:               false,
									TQLQuery:             testCase.QueryString,
									DependencyMapOptions: testCase.DependencyMapOptions,
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

func TestRunExport(t *testing.T) {
	done := make(chan struct{})

	var server *httptest.Server
	server = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/public/v0.2/blars-org/projects/blars/metric_dashboards/123abc", r.URL.Path)
		body, err := json.Marshal(map[string]client.UnifiedDashboard{
			"data": {
				ID:   "123abc",
				Type: "metric_dashboard",
				Attributes: client.UnifiedDashboardAttributes{
					Name:   "blars dashboard",
					Charts: []client.UnifiedChart{},
				},
			},
		})
		assert.NoError(t, err)
		_, err = w.Write(body)
		assert.NoError(t, err)
		w.WriteHeader(http.StatusOK)
		close(done)
	}))
	defer server.Close()

	os.Setenv("LIGHTSTEP_API_KEY", "api-key")
	os.Setenv("LIGHTSTEP_ORG", "blars-org")
	os.Setenv("LIGHTSTEP_API_URL", server.URL)
	err := Run("exporter", "export", "lightstep_dashboard", "blars", "123abc")
	assert.NoError(t, err)

	select {
	case <-time.After(10 * time.Second):
		t.Fatal("timeout waiting for HTTP handler to be exercised")
	case <-done:
	}
}
