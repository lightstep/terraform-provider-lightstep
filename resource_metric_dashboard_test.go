package main

import (
	"fmt"
	"regexp"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"

	"github.com/lightstep/terraform-provider-lightstep/lightstep"
)

func TestAccMetricDashboard(t *testing.T) {
	var dashboard lightstep.MetricDashboard

	// missing required field 'type'
	badDashboard := `
resource "lightstep_metric_dashboard" "test" {
  project_name   = "terraform-provider-tests"
  dashboard_name = "Acceptance Test Dashboard"

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

	dashboardConfig := `
resource "lightstep_metric_dashboard" "test" {
  project_name = "terraform-provider-tests"
  dashboard_name = "Acceptance Test Dashboard"

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
	updatedTitleDashboardConfig := `
resource "lightstep_metric_dashboard" "test" {
  project_name = "terraform-provider-tests"
  dashboard_name = "Acceptance Test Dashboard Updated"

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
					// TODO: verify more fields here, I don't understand how to do nested fields
				),
			},
			{
				Config: updatedTitleDashboardConfig,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckMetricDashboardExists(resourceName, &dashboard),
					resource.TestCheckResourceAttr(resourceName, "dashboard_name", "Acceptance Test Dashboard Updated"),
				),
			},
		},
	})
}

func testGetMetricDashboardDestroy(s *terraform.State) error {
	conn := testAccProvider.Meta().(*lightstep.Client)
	for _, r := range s.RootModule().Resources {
		if r.Type != "metric_alert" {
			continue
		}

		s, err := conn.GetMetricDashboard(test_project, r.Primary.ID)
		if err == nil {
			if s.ID == r.Primary.ID {
				return fmt.Errorf("Metric dashboard with ID (%v) still exists.", r.Primary.ID)
			}
		}
	}
	return nil
}

func testAccCheckMetricDashboardExists(resourceName string, dashboard *lightstep.MetricDashboard) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		tfDashboard, ok := s.RootModule().Resources[resourceName]
		if !ok {
			return fmt.Errorf("Not found: %s", resourceName)
		}

		if tfDashboard.Primary.ID == "" {
			return fmt.Errorf("ID is not set")
		}

		client := testAccProvider.Meta().(*lightstep.Client)
		dash, err := client.GetMetricDashboard(test_project, tfDashboard.Primary.ID)
		if err != nil {
			return err
		}

		*dashboard = dash
		return nil
	}
}
