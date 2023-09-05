package lightstep

import (
	"reflect"
	"strings"
	"testing"

	"github.com/lightstep/terraform-provider-lightstep/client"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

func TestAccDashboardLegacyFormat(t *testing.T) {
	var dashboard client.UnifiedDashboard

	dashboardConfig := `
resource "lightstep_metric_dashboard" "test" {
	project_name          = "` + testProject + `"
	dashboard_name        = "Acceptance Test Dashboard (TestAccDashboardLegacyFormat)"
	dashboard_description = "Dashboard to test if the legacy formats are retained when there's no diff"
	
	chart {
		name = "hit_ratio"
		rank = 1
		type = "timeseries"
	
		query {
			display             = "line"
			exclude_filters     = []
			hidden              = false
			include_filters     = []
			metric              = "cache.hit_ratio"
			query_name          = "a"
			timeseries_operator = "last"
		
			group_by {
				aggregation_method = "avg"
				keys = [
				"cache_type",
				"cache_name",
				"service",
				]
			}
		}
	}
}
`
	// Change the chart name and metric name
	updatedConfig := strings.Replace(dashboardConfig, "hit_ratio", "miss_ratio", -1)

	resourceName := "lightstep_metric_dashboard.test"

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testGetMetricDashboardDestroy,
		Steps: []resource.TestStep{
			{
				// Create the initial legacy dashboard
				Config: dashboardConfig,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckMetricDashboardExists(resourceName, &dashboard),
					resource.TestCheckResourceAttr(resourceName, "chart.0.name", "hit_ratio"),
				),
			},
			{
				// Update with no differences. Ensure the legacy format is retained.
				Config: dashboardConfig,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckMetricDashboardExists(resourceName, &dashboard),
					resource.TestCheckResourceAttr(resourceName, "chart.0.name", "hit_ratio"),
					resource.TestCheckResourceAttr(resourceName, "chart.0.rank", "1"),
					resource.TestCheckResourceAttr(resourceName, "chart.0.type", "timeseries"),
					resource.TestCheckResourceAttr(resourceName, "chart.0.query.0.tql", ""),
					resource.TestCheckResourceAttr(resourceName, "chart.0.query.0.metric", "cache.hit_ratio"),
				),
			},
			{
				// Updated config will contain the new metric and chart name
				Config: updatedConfig,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckMetricDashboardExists(resourceName, &dashboard),
					resource.TestCheckResourceAttr(resourceName, "chart.0.name", "miss_ratio"),
					resource.TestCheckResourceAttr(resourceName, "chart.0.rank", "1"),
					resource.TestCheckResourceAttr(resourceName, "chart.0.type", "timeseries"),
					resource.TestCheckResourceAttr(resourceName, "chart.0.query.0.tql", ""),
					resource.TestCheckResourceAttr(resourceName, "chart.0.query.0.metric", "cache.miss_ratio"),
				),
			},
		},
	})
}

func TestAccDashboardLegacyAndTQLFormat(t *testing.T) {
	var dashboard client.UnifiedDashboard

	dashboardConfig := `
resource "lightstep_metric_dashboard" "test" {
	project_name          = "` + testProject + `"
	dashboard_name        = "Acceptance Test Dashboard (TestAccDashboardLegacyFormat)"
	dashboard_description = "Dashboard to test if the legacy formats are retained when there's no diff"
	
	chart {
		name = "hit_ratio"
		rank = 0
		type = "timeseries"
	
		query {
			display             = "line"
			exclude_filters     = []
			hidden              = false
			include_filters     = []
			metric              = "cache.hit_ratio"
			query_name          = "a"
			timeseries_operator = "last"
		
			group_by {
				aggregation_method = "avg"
				keys = [
				"cache_type",
				"cache_name",
				"service",
				]
			}
		}
	}

	chart {
		name = "cpu"
		rank = 1
		type = "timeseries"
	
		query {
			display             = "line"
			hidden              = false
			query_name          = "a"		
			tql					= "metric cpu.utilization | latest | group_by [], sum"
		}
	}
}
`
	// Change the chart name and metric name
	updatedConfig1 := strings.Replace(dashboardConfig, "hit_ratio", "miss_ratio", -1)
	// Update the TQL query
	updatedConfig2 := strings.Replace(updatedConfig1, "group_by [], sum", "group_by [], mean", -1)

	resourceName := "lightstep_metric_dashboard.test"

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: testAccProviderFactories,
		CheckDestroy:      testGetMetricDashboardDestroy,
		Steps: []resource.TestStep{
			{
				// Create the initial legacy dashboard
				Config: dashboardConfig,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckMetricDashboardExists(resourceName, &dashboard),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "chart.*",
						map[string]string{
							"name":           "hit_ratio",
							"query.0.metric": "cache.hit_ratio",
							"query.0.tql":    "",
							"rank":           "0",
							"type":           "timeseries",
						},
					),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "chart.*",
						map[string]string{
							"name":        "cpu",
							"query.0.tql": "metric cpu.utilization | latest | group_by [], sum",
						},
					),
				),
			},
			{
				// Update with no differences. Ensure the legacy format and TQL are retained.
				Config: dashboardConfig,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckMetricDashboardExists(resourceName, &dashboard),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "chart.*",
						map[string]string{
							"name":           "hit_ratio",
							"query.0.metric": "cache.hit_ratio",
							"query.0.tql":    "",
							"rank":           "0",
							"type":           "timeseries",
						},
					),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "chart.*",
						map[string]string{
							"name":        "cpu",
							"query.0.tql": "metric cpu.utilization | latest | group_by [], sum",
						},
					),
				),
			},
			{
				// Updated config will contain the new metric and chart name in chart 0
				Config: updatedConfig1,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckMetricDashboardExists(resourceName, &dashboard),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "chart.*",
						map[string]string{
							"name":        "miss_ratio",
							"query.0.tql": "", // Should still be legacy
						},
					),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "chart.*",
						map[string]string{
							"name":        "cpu",
							"query.0.tql": "metric cpu.utilization | latest | group_by [], sum",
						},
					),
				),
			},
			{
				// Updated config will the TQL query of chart 1
				Config: updatedConfig2,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckMetricDashboardExists(resourceName, &dashboard),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "chart.*",
						map[string]string{
							"name":        "miss_ratio",
							"query.0.tql": "",
						},
					),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "chart.*",
						map[string]string{
							"name":        "cpu",
							"query.0.tql": "metric cpu.utilization | latest | group_by [], mean",
						},
					),
				),
			},
		},
	})
}

func TestAccDashboardVPADashTest(t *testing.T) {
	var dashboard client.UnifiedDashboard

	dashboardConfig := `
resource "lightstep_metric_dashboard" "test" {
 project_name   = "` + testProject + `"
 dashboard_name = "VPA (VerticalPodAutoscaler) - TimeSeries (terraform)"

 chart {
   name = "CPU: Capped Target"
   description = "VPA will adjust pods' CPU requests until the average across all replicas is less than or equal to this value"
   rank = "0"
   type = "timeseries"

   query {
     query_name = "a"
     display    = "line"
     hidden     = false

     metric              = "kubernetes_state.vpa.target"
     timeseries_operator = "last"

     include_filters = [
       {
         key   = "resource"
         value = "cpu"
       },
     ]


     group_by {
       aggregation_method = "sum"
       keys               = []
     }

   }

 }

 chart {
   name = "Memory: Capped Target"
   rank = "1"
   type = "timeseries"

   query {
     query_name = "a"
     display    = "line"
     hidden     = false

     metric              = "kubernetes_state.vpa.target"
     timeseries_operator = "last"

     include_filters = [
       {
         key   = "resource"
         value = "memory"
       },
     ]


     group_by {
       aggregation_method = "sum"
       keys               = []
     }

   }

 }

 chart {
   name = "CPU: Uncapped Target"
   rank = "3"
   type = "timeseries"

   query {
     query_name = "a"
     display    = "line"
     hidden     = false

     metric              = "kubernetes_state.vpa.uncapped_target"
     timeseries_operator = "last"

     include_filters = [
       {
         key   = "resource"
         value = "cpu"
       },
     ]


     group_by {
       aggregation_method = "sum"
       keys               = []
     }

   }

 }

 chart {
   name = "Memory: Uncapped Target"
   rank = "4"
   type = "timeseries"

   query {
     query_name = "a"
     display    = "line"
     hidden     = false

     metric              = "kubernetes_state.vpa.uncapped_target"
     timeseries_operator = "last"

     include_filters = [
       {
         key   = "resource"
         value = "memory"
       },
     ]


     group_by {
       aggregation_method = "sum"
       keys               = []
     }

   }

 }

}
`

	resourceName := "lightstep_metric_dashboard.test"

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testGetMetricDashboardDestroy,
		Steps: []resource.TestStep{
			{
				// Create the initial legacy dashboard
				Config: dashboardConfig,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckMetricDashboardExists(resourceName, &dashboard),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "chart.*", map[string]string{
						"name":        "CPU: Capped Target",
						"description": "VPA will adjust pods' CPU requests until the average across all replicas is less than or equal to this value",
					}),
				),
			},
			{
				// Update with no differences. Ensure the legacy format is retained.
				Config: dashboardConfig,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckMetricDashboardExists(resourceName, &dashboard),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "chart.*", map[string]string{
						"name":        "CPU: Capped Target",
						"description": "VPA will adjust pods' CPU requests until the average across all replicas is less than or equal to this value",
					}),
				),
			},
		},
	})
}

func Test_buildLabels(t *testing.T) {
	tests := []struct {
		name        string
		in          []interface{}
		want        []client.Label
		wantErr     bool
		lengthCheck bool
	}{
		{
			name:    "empty label",
			in:      []interface{}{map[string]interface{}{}},
			want:    nil,
			wantErr: false,
		},
		{
			name: "basic label",
			in: []interface{}{map[string]interface{}{
				"key":   "team",
				"value": "ontology",
			}},
			want:    []client.Label{{Key: "team", Value: "ontology"}},
			wantErr: false,
		},
		{
			name: "two basic labels",
			in: []interface{}{map[string]interface{}{
				"key":   "team",
				"value": "ontology",
			}, map[string]interface{}{
				"key":   "env",
				"value": "meta",
			}},
			want: []client.Label{
				{Key: "team", Value: "ontology"},
				{Key: "env", Value: "meta"},
			},
			wantErr: false,
		},
		{
			name: "label without key returns just the value",
			in: []interface{}{map[string]interface{}{
				"value": "ontology",
			}},
			want:    []client.Label{{Key: "", Value: "ontology"}},
			wantErr: false,
		},
		{
			name: "label key must be string",
			in: []interface{}{map[string]interface{}{
				"key":   2,
				"value": "ontology",
			}},
			want:    nil,
			wantErr: true,
		},
		{
			name: "label value required",
			in: []interface{}{map[string]interface{}{
				"key": "test",
			}},
			want:    nil,
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := buildLabels(tt.in)
			if (err != nil) != tt.wantErr {
				t.Errorf("buildLabels() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("buildLabels() got = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_extractLabels(t *testing.T) {
	tests := []struct {
		name      string
		apiLabels []client.Label
		want      []interface{}
	}{
		{
			name:      "empty labels",
			apiLabels: []client.Label{},
			want:      nil,
		},
		{
			name:      "basic label",
			apiLabels: []client.Label{{Key: "team", Value: "ontology"}},
			want: []interface{}{map[string]interface{}{
				"key":   "team",
				"value": "ontology",
			}},
		},
		{
			name:      "basic label without key",
			apiLabels: []client.Label{{Value: "ontology"}},
			want: []interface{}{map[string]interface{}{
				"value": "ontology",
			}},
		},
		{
			name:      "basic labels",
			apiLabels: []client.Label{{Value: "ontology"}, {Value: "meta"}},
			want: []interface{}{map[string]interface{}{
				"value": "ontology",
			}, map[string]interface{}{
				"value": "meta",
			}},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := extractLabels(tt.apiLabels); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("extractLabels() = %v, want %v", got, tt.want)
			}
		})
	}
}
