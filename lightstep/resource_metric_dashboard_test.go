package lightstep

import (
	"strings"
	"context"
	"fmt"
	"regexp"
	"testing"

	"github.com/lightstep/terraform-provider-lightstep/client"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
)

func TestAccMetricDashboard(t *testing.T) {
	var dashboard client.UnifiedDashboard

	// missing required field 'type'
	badDashboard := `
resource "lightstep_metric_dashboard" "test" {
  project_name          = "terraform-provider-tests"
  dashboard_name        = "Acceptance Test Dashboard"
  dashboard_description = "Dashboard to test if the terraform provider works"

  chart {
    name = "Chart Number One"
    rank = 1

    y_axis {
      min = 0.4
      max = 5.0
    }


    query {
      hidden              = false
      query_name          = "a"
      display             = "bar"
      timeseries_operator = "rate"
      metric              = "pagerduty.task.success"

      include_filters = [{
        key   = "kube_app"
        value = "pagerduty"
      }]

      group_by {
        aggregation_method = "max"
        keys               = ["cluster-name"]
      }
    }
  }
}
`

	tqlDashboardConfig := `
resource "lightstep_metric_dashboard" "test" {
  project_name          = "terraform-provider-tests"
  dashboard_name        = "Acceptance Test Dashboard"
  dashboard_description = "Dashboard to test if the terraform provider works"

  chart {
    name = "Chart Number One"
    rank = 1
    type = "timeseries"

    query {
      hidden              = false
      query_name          = "a"
      display             = "line"
      tql                 = "metric m | rate"
    }
  }
}`

	groupedDashboardConfig := `
	resource "lightstep_metric_dashboard" "test_groups" {
	 project_name          = "terraform-provider-tests"
	 dashboard_name        = "Acceptance Test Dashboard"
	 dashboard_description = "Dashboard to test if the terraform provider works"
	 group {
		title = "Title"
		rank = 0
		visibility_type = "explicit"
		chart {
		  name = "Chart Number One"
		  rank = 1
		  type = "timeseries"
	
		  query {
		    hidden              = false
		    query_name          = "a"
		    display             = "line"
		    tql                 = "metric m | rate"
		  }
		}
	 }
	}
	`
	spansQueryDashboardConfig := `
resource "lightstep_metric_dashboard" "test_spans" {
  project_name          = "terraform-provider-tests"
  dashboard_name        = "Acceptance Test Dashboard"
  dashboard_description = "Dashboard to test if the terraform provider works"

  chart {
    name = "Chart Number One"
    rank = 1
    type = "timeseries"

    query {
      hidden              = false
      query_name          = "a"
      display             = "line"

      spans {
        query = "service IN (\"frontend\")"
        operator = "error_ratio"
        operator_input_window_ms = 3600000
      }

      final_window_operation {
        operator = "min"
        input_window_ms  = 30000
      }
    }
  }
}
`

	dashboardConfig := `
resource "lightstep_metric_dashboard" "test" {
  project_name          = "terraform-provider-tests"
  dashboard_name        = "Acceptance Test Dashboard"
  dashboard_description = "Dashboard to test if the terraform provider works"

  chart {
    name = "Chart Number One"
    rank = 1
    type = "timeseries"

    y_axis {
      min = 0.4
      max = 5.0
    }


    query {
      hidden              = false
      query_name          = "a"
      display             = "bar"
      timeseries_operator = "rate"
      metric              = "pagerduty.task.success"

      include_filters = [{
        key   = "kube_app"
        value = "pagerduty"
      }]

      filters = [{
        key   = "kube_app"
        operand = "contains"
        value = "frontend"
      }]

      group_by {
        aggregation_method = "max"
        keys               = ["cluster-name"]
      }
    }
  }
}
`
	updatedTitleDashboardConfig := `
resource "lightstep_metric_dashboard" "test" {
  project_name          = "terraform-provider-tests"
  dashboard_name        = "Acceptance Test Dashboard Updated"
  dashboard_description = "Dashboard to test if the terraform provider still works"

  chart {
    name = "Chart Number One"
    rank = 1
    type = "timeseries"

    y_axis {
      min = 0.4
      max = 5.0
    }


    query {
      hidden              = false
      query_name          = "a"
      display             = "bar"
      timeseries_operator = "rate"
      metric              = "pagerduty.task.success"

      include_filters = [{
        key   = "kube_app"
        value = "pagerduty"
      }]

      group_by {
        aggregation_method = "max"
        keys               = ["cluster-name"]
      }
    }
  }
}
`

	resourceName := "lightstep_metric_dashboard.test"
	resourceNameSpans := "lightstep_metric_dashboard.test_spans"
)

func TestAccDashboardLegacyFormat(t *testing.T) {
	var dashboard client.UnifiedDashboard

	dashboardConfig := `
resource "lightstep_metric_dashboard" "test" {
	project_name          = "terraform-provider-tests"
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
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: testAccProviderFactories,
		CheckDestroy:      testGetMetricDashboardDestroy,
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
	project_name          = "terraform-provider-tests"
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
				Config: badDashboard,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckMetricDashboardExists(resourceName, &dashboard)),
				ExpectError: regexp.MustCompile("The argument \"type\" is required, but no definition was found."),
			},
			{
				Config: dashboardConfig,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckMetricDashboardExists(resourceName, &dashboard),
					resource.TestCheckResourceAttr(resourceName, "dashboard_name", "Acceptance Test Dashboard"),
					resource.TestCheckResourceAttr(resourceName, "dashboard_description", "Dashboard to test if the terraform provider works"),
					resource.TestCheckResourceAttr(resourceName, "chart.0.name", "Chart Number One"),
					resource.TestCheckResourceAttr(resourceName, "chart.0.rank", "1"),
					resource.TestCheckResourceAttr(resourceName, "chart.0.type", "timeseries"),
					resource.TestCheckResourceAttr(resourceName, "chart.0.query.0.metric", "pagerduty.task.success"),
					resource.TestCheckResourceAttr(resourceName, "chart.0.query.0.timeseries_operator", "rate"),
					resource.TestCheckResourceAttr(resourceName, "chart.0.query.0.timeseries_operator_input_window_ms", "0"),
					resource.TestCheckResourceAttr(resourceName, "chart.0.query.0.display", "bar"),
					resource.TestCheckResourceAttr(resourceName, "chart.0.query.0.hidden", "false"),
					resource.TestCheckResourceAttr(resourceName, "chart.0.query.0.include_filters.0.key", "kube_app"),
					resource.TestCheckResourceAttr(resourceName, "chart.0.query.0.include_filters.0.value", "pagerduty"),
					resource.TestCheckResourceAttr(resourceName, "chart.0.query.0.filters.0.key", "kube_app"),
					resource.TestCheckResourceAttr(resourceName, "chart.0.query.0.filters.0.operand", "contains"),
					resource.TestCheckResourceAttr(resourceName, "chart.0.query.0.filters.0.value", "frontend"),
				),
			},
			{
				Config: tqlDashboardConfig,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckMetricDashboardExists(resourceName, &dashboard),
					resource.TestCheckResourceAttr(resourceName, "dashboard_name", "Acceptance Test Dashboard"),
					resource.TestCheckResourceAttr(resourceName, "dashboard_description", "Dashboard to test if the terraform provider works"),
					resource.TestCheckResourceAttr(resourceName, "chart.0.query.0.tql", "metric m | rate"),
					resource.TestCheckResourceAttr(resourceName, "chart.0.query.0.display", "line"),
					resource.TestCheckResourceAttr(resourceName, "chart.0.query.0.hidden", "false"),
				),
			},
			{
				Config: groupedDashboardConfig,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckMetricDashboardExists(resourceName, &dashboard),
					resource.TestCheckResourceAttr(resourceName, "dashboard_name", "Acceptance Test Dashboard"),
					resource.TestCheckResourceAttr(resourceName, "dashboard_description", "Dashboard to test if the terraform provider works"),
					resource.TestCheckResourceAttr(resourceName, "group.0.title", "Title"),
					resource.TestCheckResourceAttr(resourceName, "group.0.rank", "0"),
					resource.TestCheckResourceAttr(resourceName, "group.0.visibility_type", "explicit"),
					resource.TestCheckResourceAttr(resourceName, "group.0.chart.0.query.0.tql", "metric m | rate"),
					resource.TestCheckResourceAttr(resourceName, "group.0.chart.0.query.0.display", "line"),
					resource.TestCheckResourceAttr(resourceName, "group.0.chart.0.query.0.hidden", "false"),
				),
			},
			{
				Config: spansQueryDashboardConfig,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckMetricDashboardExists(resourceNameSpans, &dashboard),
					resource.TestCheckResourceAttr(resourceNameSpans, "dashboard_name", "Acceptance Test Dashboard"),
					resource.TestCheckResourceAttr(resourceNameSpans, "dashboard_description", "Dashboard to test if the terraform provider works"),
					resource.TestCheckResourceAttr(resourceNameSpans, "chart.0.query.0.display", "line"),
					resource.TestCheckResourceAttr(resourceNameSpans, "chart.0.query.0.hidden", "false"),
					resource.TestCheckResourceAttr(resourceNameSpans, "chart.0.query.0.spans.0.operator", "error_ratio"),
					resource.TestCheckResourceAttr(resourceNameSpans, "chart.0.query.0.spans.0.operator_input_window_ms", "3600000"),
					resource.TestCheckResourceAttr(resourceNameSpans, "chart.0.query.0.spans.0.query", "service IN (\"frontend\")"),
					resource.TestCheckResourceAttr(resourceNameSpans, "chart.0.query.0.final_window_operation.0.operator", "min"),
					resource.TestCheckResourceAttr(resourceNameSpans, "chart.0.query.0.final_window_operation.0.input_window_ms", "30000"),
				),
			},
			{
				Config: updatedTitleDashboardConfig,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckMetricDashboardExists(resourceName, &dashboard),
					resource.TestCheckResourceAttr(resourceName, "dashboard_name", "Acceptance Test Dashboard Updated"),
					resource.TestCheckResourceAttr(resourceName, "dashboard_description", "Dashboard to test if the terraform provider still works"),
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

func testGetMetricDashboardDestroy(s *terraform.State) error {
	conn := testAccProvider.Meta().(*client.Client)
	for _, r := range s.RootModule().Resources {
		if r.Type != "metric_alert" {
			continue
		}

		s, err := conn.GetUnifiedDashboard(context.Background(), test_project, r.Primary.ID, false)
		if err == nil {
			if s.ID == r.Primary.ID {
				return fmt.Errorf("metric dashboard with ID (%v) still exists.", r.Primary.ID)
			}
		}
	}
	return nil
}

func testAccCheckMetricDashboardExists(resourceName string, dashboard *client.UnifiedDashboard) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		tfDashboard, ok := s.RootModule().Resources[resourceName]
		if !ok {
			return fmt.Errorf("not found: %s", resourceName)
		}

		if tfDashboard.Primary.ID == "" {
			return fmt.Errorf("id is not set")
		}

		c := testAccProvider.Meta().(*client.Client)
		dash, err := c.GetUnifiedDashboard(context.Background(), test_project, tfDashboard.Primary.ID, false)
		if err != nil {
			return err
		}

		dashboard = dash
		return nil
	}
func TestAccDashboardVPADashTest(t *testing.T) {
	var dashboard client.UnifiedDashboard

	dashboardConfig := `
resource "lightstep_metric_dashboard" "test" {
 project_name   = "terraform-provider-tests"
 dashboard_name = "VPA (VerticalPodAutoscaler) - TimeSeries (terraform)"

 chart {
   name = "CPU: Capped Target"
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
						"name": "CPU: Capped Target",
					}),
				),
			},
			{
				// Update with no differences. Ensure the legacy format is retained.
				Config: dashboardConfig,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckMetricDashboardExists(resourceName, &dashboard),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "chart.*", map[string]string{
						"name": "CPU: Capped Target",
					}),
				),
			},
		},
	})
}
