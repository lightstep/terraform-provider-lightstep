package lightstep

import (
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
	templateVariableDashboardConfig := `
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
      tql                 = "metric m | filter (service == $service) | rate | group_by [], sum"
    }
  }

  template_variable {
	name = "service"
	default_values = ["myService"]
    suggestion_attribute_key = "service.name"
  }
}
`

	resourceName := "lightstep_metric_dashboard.test"
	resourceNameSpans := "lightstep_metric_dashboard.test_spans"

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
				),
			},
			{
				Config: templateVariableDashboardConfig,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckMetricDashboardExists(resourceName, &dashboard),
					resource.TestCheckResourceAttr(resourceName, "dashboard_name", "Acceptance Test Dashboard"),
					resource.TestCheckResourceAttr(resourceName, "dashboard_description", "Dashboard to test if the terraform provider works"),
					resource.TestCheckResourceAttr(resourceName, "chart.0.query.0.tql", "metric m | filter (service == $service) | rate | group_by [], sum"),
					resource.TestCheckResourceAttr(resourceName, "chart.0.query.0.display", "line"),
					resource.TestCheckResourceAttr(resourceName, "chart.0.query.0.hidden", "false"),
					resource.TestCheckResourceAttr(resourceName, "template_variable.0.name", "service"),
					resource.TestCheckResourceAttr(resourceName, "template_variable.0.default_values.0", "myService"),
					resource.TestCheckResourceAttr(resourceName, "template_variable.0.suggestion_attribute_key", "service.name"),
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
}
