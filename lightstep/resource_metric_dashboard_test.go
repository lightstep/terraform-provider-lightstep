package lightstep

import (
	"strings"
	"testing"

	"github.com/lightstep/terraform-provider-lightstep/client"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
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
	t.Skip("Known issue with chart order not being consistent")

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
				// Create the initial legacy dashboard
				Config: dashboardConfig,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckMetricDashboardExists(resourceName, &dashboard),
					resource.TestCheckResourceAttr(resourceName, "chart.1.name", "hit_ratio"),
					resource.TestCheckResourceAttr(resourceName, "chart.1.rank", "0"),
					resource.TestCheckResourceAttr(resourceName, "chart.1.type", "timeseries"),
					resource.TestCheckResourceAttr(resourceName, "chart.1.query.0.tql", ""),
					resource.TestCheckResourceAttr(resourceName, "chart.1.query.0.metric", "cache.hit_ratio"),

					resource.TestCheckResourceAttr(resourceName, "chart.0.name", "cpu"),
					resource.TestCheckResourceAttr(resourceName, "chart.0.query.0.tql", "metric cpu.utilization | latest | group_by [], sum"),
				),
			},
			{
				// Update with no differences. Ensure the legacy format and TQL are retained.
				Config: dashboardConfig,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckMetricDashboardExists(resourceName, &dashboard),
					resource.TestCheckResourceAttr(resourceName, "chart.1.name", "hit_ratio"),
					resource.TestCheckResourceAttr(resourceName, "chart.1.rank", "0"),
					resource.TestCheckResourceAttr(resourceName, "chart.1.type", "timeseries"),
					resource.TestCheckResourceAttr(resourceName, "chart.1.query.0.tql", ""),
					resource.TestCheckResourceAttr(resourceName, "chart.1.query.0.metric", "cache.hit_ratio"),

					resource.TestCheckResourceAttr(resourceName, "chart.0.name", "cpu"),
					resource.TestCheckResourceAttr(resourceName, "chart.0.query.0.tql", "metric cpu.utilization | latest | group_by [], sum"),
				),
			},
			{
				// Updated config will contain the new metric and chart name in chart 0
				Config: updatedConfig1,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckMetricDashboardExists(resourceName, &dashboard),
					resource.TestCheckResourceAttr(resourceName, "chart.1.name", "miss_ratio"),
					resource.TestCheckResourceAttr(resourceName, "chart.1.rank", "0"),
					resource.TestCheckResourceAttr(resourceName, "chart.1.type", "timeseries"),
					resource.TestCheckResourceAttr(resourceName, "chart.1.query.0.tql", ""),
					resource.TestCheckResourceAttr(resourceName, "chart.1.query.0.metric", "cache.miss_ratio"),

					resource.TestCheckResourceAttr(resourceName, "chart.0.name", "cpu"),
					resource.TestCheckResourceAttr(resourceName, "chart.0.query.0.tql", "metric cpu.utilization | latest | group_by [], sum"),
				),
				ExpectNonEmptyPlan: true,
			},
			{
				// Updated config will update only the TQL query of chart 1
				Config: updatedConfig2,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckMetricDashboardExists(resourceName, &dashboard),
					resource.TestCheckResourceAttr(resourceName, "chart.1.name", "miss_ratio"),
					resource.TestCheckResourceAttr(resourceName, "chart.1.rank", "0"),
					resource.TestCheckResourceAttr(resourceName, "chart.1.type", "timeseries"),
					resource.TestCheckResourceAttr(resourceName, "chart.1.query.0.tql", ""),
					resource.TestCheckResourceAttr(resourceName, "chart.1.query.0.metric", "cache.miss_ratio"),

					resource.TestCheckResourceAttr(resourceName, "chart.0.name", "cpu"),
					resource.TestCheckResourceAttr(resourceName, "chart.0.query.0.tql", "metric cpu.utilization | latest | group_by [], mean"),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}
