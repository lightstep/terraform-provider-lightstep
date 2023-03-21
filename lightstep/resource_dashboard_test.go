package lightstep

import (
	"context"
	"fmt"
	"regexp"
	"strings"
	"testing"

	"github.com/lightstep/terraform-provider-lightstep/client"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
)

func TestAccDashboard(t *testing.T) {
	var dashboard client.UnifiedDashboard

	// missing required field 'type'
	badDashboard := `
	resource "lightstep_dashboard" "test" {
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
	     query_string        = "metric requests | rate 10m"
	   }
	 }
	}
	`

	queryDashboardConfig := `
	resource "lightstep_dashboard" "test" {
	 project_name = "terraform-provider-tests"
	 dashboard_name = "Acceptance Test Dashboard"
	
	 chart {
	   name = "Chart Number One"
	   rank = 1
	   type = "timeseries"
	
	   query {
	     hidden              = false
	     query_name          = "a"
	     display             = "line"
	     query_string        = "metric m | rate"
	   }
	 }
	}
	`

	dashboardConfig := `
	resource "lightstep_dashboard" "test" {
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
	     query_string        = "metric pagerduty.task.success | rate 10m | filter kube_app = \"pagerduty\" | group_by[\"cluster-name\"], max"
	   }
	 }
	}
	`
	updatedTitleDashboardConfig := `
	resource "lightstep_dashboard" "test" {
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
	     query_string        = "metric pagerduty.task.success | rate 10m | filter kube_app = \"pagerduty\" | group_by[\"cluster-name\"], max"
	   }
	 }
	}
	`

	dependencyMapDashboard := `
	resource "lightstep_dashboard" "test" {
	 project_name   = "terraform-provider-tests"
	 dashboard_name = "Acceptance Test Dashboard"
	
	 chart {
	   name = "Chart Number One"
	   rank = 1
	   type = "timeseries"
	
	   query {
	     hidden       = false
	     query_name   = "a"
	     display      = "dependency_map"
	     query_string = "spans_sample service = apache | assemble"
	     dependency_map_options {
	       scope    = "upstream"
	       map_type = "operation"
	     }
	   }
	 }
	}
	`

	groupedDashboardConfig := `
	resource "lightstep_dashboard" "test" {
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
				query_string        = "metric m | rate"
			  }
			}
		}
	}
	`

	positionallyGroupedImplicitDashboardConfig := `
resource "lightstep_dashboard" "test" {
  dashboard_name = "Acceptance Test Dashboard"
  dashboard_description = "Dashboard to test if the terraform provider works"
  project_name   = "terraform-provider-tests"

  group {
    rank            = 0
    title           = ""
    visibility_type = "implicit"

    chart {
      name   = "responses"
      type   = "timeseries"
      rank   = 0
      x_pos  = 0
      y_pos  = 0
      width  = 32
      height = 10

      query {
        query_name   = "a"
        display      = "line"
        hidden       = false
        query_string = "metric responses | rate | group_by [], sum"
      }
    }
  }
}
`
	positionallyGroupedExplicitDashboardConfig := `
resource "lightstep_dashboard" "test" {
  dashboard_name = "Acceptance Test Dashboard"
  dashboard_description = "Dashboard to test if the terraform provider works"
  project_name   = "terraform-provider-tests"

  group {
    rank            = 0
    title           = "Title"
    visibility_type = "explicit"

    chart {
      name   = "responses"
      type   = "timeseries"
      rank   = 0
      x_pos  = 16
      y_pos  = 6
      width  = 32
      height = 10

      query {
        query_name   = "a"
        display      = "line"
        hidden       = false
        query_string = "metric responses | rate | group_by [], sum"
      }
    }
  }
}
`

	resourceName := "lightstep_dashboard.test"

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testGetMetricDashboardDestroy,
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
					resource.TestCheckResourceAttr(resourceName, "chart.0.name", "Chart Number One"),
					resource.TestCheckResourceAttr(resourceName, "chart.0.rank", "1"),
					resource.TestCheckResourceAttr(resourceName, "chart.0.type", "timeseries"),
					resource.TestCheckResourceAttr(resourceName, "chart.0.query.0.display", "bar"),
					resource.TestCheckResourceAttr(resourceName, "chart.0.query.0.hidden", "false"),
				),
			},
			{
				Config: queryDashboardConfig,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckMetricDashboardExists(resourceName, &dashboard),
					resource.TestCheckResourceAttr(resourceName, "dashboard_name", "Acceptance Test Dashboard"),
					resource.TestCheckResourceAttr(resourceName, "chart.0.query.0.query_string", "metric m | rate"),
					resource.TestCheckResourceAttr(resourceName, "chart.0.query.0.display", "line"),
					resource.TestCheckResourceAttr(resourceName, "chart.0.query.0.hidden", "false"),
				),
			},
			{
				Config: updatedTitleDashboardConfig,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckMetricDashboardExists(resourceName, &dashboard),
					resource.TestCheckResourceAttr(resourceName, "dashboard_name", "Acceptance Test Dashboard Updated"),
				),
			},
			{
				Config: dependencyMapDashboard,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckMetricDashboardExists(resourceName, &dashboard),
					resource.TestCheckResourceAttr(resourceName, "dashboard_name", "Acceptance Test Dashboard"),
					resource.TestCheckResourceAttr(resourceName, "chart.0.query.0.query_string", "spans_sample service = apache | assemble"),
					resource.TestCheckResourceAttr(resourceName, "chart.0.query.0.display", "dependency_map"),
					resource.TestCheckResourceAttr(resourceName, "chart.0.query.0.hidden", "false"),
					resource.TestCheckResourceAttr(resourceName, "chart.0.query.0.dependency_map_options.0.scope", "upstream"),
					resource.TestCheckResourceAttr(resourceName, "chart.0.query.0.dependency_map_options.0.map_type", "operation"),
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
					resource.TestCheckResourceAttr(resourceName, "group.0.chart.0.query.0.query_string", "metric m | rate"),
					resource.TestCheckResourceAttr(resourceName, "group.0.chart.0.query.0.display", "line"),
					resource.TestCheckResourceAttr(resourceName, "group.0.chart.0.query.0.hidden", "false"),
				),
			},
			{
				Config: positionallyGroupedImplicitDashboardConfig,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckMetricDashboardExists(resourceName, &dashboard),
					resource.TestCheckResourceAttr(resourceName, "dashboard_name", "Acceptance Test Dashboard"),
					resource.TestCheckResourceAttr(resourceName, "dashboard_description", "Dashboard to test if the terraform provider works"),
					resource.TestCheckResourceAttr(resourceName, "group.0.title", ""),
					resource.TestCheckResourceAttr(resourceName, "group.0.rank", "0"),
					resource.TestCheckResourceAttr(resourceName, "group.0.visibility_type", "implicit"),
					resource.TestCheckResourceAttr(resourceName, "group.0.chart.0.x_pos", "0"),
					resource.TestCheckResourceAttr(resourceName, "group.0.chart.0.y_pos", "0"),
					resource.TestCheckResourceAttr(resourceName, "group.0.chart.0.width", "32"),
					resource.TestCheckResourceAttr(resourceName, "group.0.chart.0.height", "10"),
				),
			},
			{
				Config: positionallyGroupedExplicitDashboardConfig,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckMetricDashboardExists(resourceName, &dashboard),
					resource.TestCheckResourceAttr(resourceName, "dashboard_name", "Acceptance Test Dashboard"),
					resource.TestCheckResourceAttr(resourceName, "dashboard_description", "Dashboard to test if the terraform provider works"),
					resource.TestCheckResourceAttr(resourceName, "group.0.title", "Title"),
					resource.TestCheckResourceAttr(resourceName, "group.0.rank", "0"),
					resource.TestCheckResourceAttr(resourceName, "group.0.visibility_type", "explicit"),
					resource.TestCheckResourceAttr(resourceName, "group.0.chart.0.x_pos", "16"),
					resource.TestCheckResourceAttr(resourceName, "group.0.chart.0.y_pos", "6"),
					resource.TestCheckResourceAttr(resourceName, "group.0.chart.0.width", "32"),
					resource.TestCheckResourceAttr(resourceName, "group.0.chart.0.height", "10"),
				),
			},
		},
	})
}

func TestAccDashboard2(t *testing.T) {
	var dashboard client.UnifiedDashboard

	uqlQuery := `metric pagerduty.task.success | filter (kube_app == "pagerduty") | rate | group_by ["cluster-name"], max`

	// missing required field 'type'
	badDashboard := fmt.Sprintf(`
resource "lightstep_dashboard" "test" {
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
      query_string        = <<EOT
%s
EOT
    }
  }
}
`, uqlQuery)

	dashboardConfig := fmt.Sprintf(`
resource "lightstep_dashboard" "test" {
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
      query_string        = <<EOT
%s
EOT
    }
  }
}
`, uqlQuery)

	updatedTitleDashboardConfig := fmt.Sprintf(`
resource "lightstep_dashboard" "test" {
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
      query_string        = <<EOT
%s
EOT
    }
  }
}
`, uqlQuery)

	resourceName := "lightstep_dashboard.test"

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
					resource.TestCheckResourceAttr(resourceName, "chart.0.query.0.query_string", uqlQuery+"\n"),
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
		},
	})
}

func TestAccDashboardLabels(t *testing.T) {
	var dashboard client.UnifiedDashboard

	baseConfig := `
resource "lightstep_dashboard" "labels" {
  project_name   = "terraform-provider-tests"
  dashboard_name = "Acceptance Test Dashboard"

  label {
    key = "team"
    value = "ontology"
  }

  chart {
    name = "Chart Number One"
    rank = 1
    type = "timeseries"

    query {
      hidden              = false
      query_name          = "a"
      display             = "bar"
      query_string        = "metric requests | rate 10m"
    }
  }
}
`

	updatedConfig := `
resource "lightstep_dashboard" "labels" {
  project_name   = "terraform-provider-tests"
  dashboard_name = "Acceptance Test Dashboard"

  chart {
    name = "Chart Number One"
    rank = 1
    type = "timeseries"

    query {
      hidden              = false
      query_name          = "a"
      display             = "bar"
      query_string        = "metric requests | rate 10m"
    }
  }
}
	`

	resourceName := "lightstep_dashboard.labels"

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: testAccProviderFactories,
		CheckDestroy:      testGetMetricDashboardDestroy,
		Steps: []resource.TestStep{
			{
				Config: baseConfig,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckMetricDashboardExists(resourceName, &dashboard),
					resource.TestCheckResourceAttr(resourceName, "label.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "label.0.key", "team"),
					resource.TestCheckResourceAttr(resourceName, "label.0.value", "ontology"),
				),
			},
			{
				Config: updatedConfig,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckMetricDashboardExists(resourceName, &dashboard),
					resource.TestCheckResourceAttr(resourceName, "label.#", "0"),
				),
			},
		},
	})
}

func TestAccDashboardWithTemplateVariables(t *testing.T) {
	validDashboardConfigWithTemplateVariables := `
resource "lightstep_dashboard" "test" {
  project_name          = "terraform-provider-tests"
  dashboard_name        = "Acceptance Test Dashboard with Template Variables"
  chart {
    name = "Chart Number One"
    rank = 1
    type = "timeseries"
    query {
      hidden              = false
      query_name          = "a"
      display             = "line"
      query_string        = "metric m | filter (service == $service) | rate | group_by [], sum"
    }
  }
  template_variable {
	name = "service"
	default_values = ["myService"]
    suggestion_attribute_key = "service_name"
  }
}
`

	invalidDashboardConfigWithInvalidTemplateVariableName := `
resource "lightstep_dashboard" "test" {
  project_name          = "terraform-provider-tests"
  dashboard_name        = "Acceptance Test Dashboard"
  chart {
    name = "Chart Number One"
    rank = 1
    type = "timeseries"
    query {
      hidden              = false
      query_name          = "a"
      display             = "line"
      query_string        = "metric m | filter (service == $invalid-name) | rate | group_by [], sum"
    }
  }
  template_variable {
	name = "invalid-name"
	default_values = ["myService"]
    suggestion_attribute_key = "service_name"
  }
}
`

	var dashboard client.UnifiedDashboard
	resourceName := "lightstep_dashboard.test"

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: testAccProviderFactories,
		CheckDestroy:      testGetMetricDashboardDestroy,
		Steps: []resource.TestStep{
			{
				Config: validDashboardConfigWithTemplateVariables,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckMetricDashboardExists(resourceName, &dashboard),
					resource.TestCheckResourceAttr(resourceName, "dashboard_name", "Acceptance Test Dashboard with Template Variables"),
					resource.TestCheckResourceAttr(resourceName, "chart.0.query.0.query_string", "metric m | filter (service == $service) | rate | group_by [], sum"),
					resource.TestCheckResourceAttr(resourceName, "chart.0.query.0.display", "line"),
					resource.TestCheckResourceAttr(resourceName, "chart.0.query.0.hidden", "false"),
					resource.TestCheckResourceAttr(resourceName, "template_variable.0.name", "service"),
					resource.TestCheckResourceAttr(resourceName, "template_variable.0.default_values.0", "myService"),
					resource.TestCheckResourceAttr(resourceName, "template_variable.0.suggestion_attribute_key", "service_name"),
				),
			},
			{
				Config: invalidDashboardConfigWithInvalidTemplateVariableName,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckMetricDashboardExists(resourceName, &dashboard)),
				ExpectError: regexp.MustCompile("InvalidArgument"),
			},
		},
	})
}

func TestAccDashboardHiddenQueries(t *testing.T) {

	queryString := strings.TrimSpace(`
with
	b = metric m | filter (service == "cats") | rate | group_by [], sum;
	c = metric m | filter (service == "dogs") | rate | group_by [], sum;
join b + c, b = 0, c = 0
`)

	config := fmt.Sprintf(`
resource "lightstep_dashboard" "test" {
	project_name          = "terraform-provider-tests"
	dashboard_name        = "Acceptance Test Dashboard with Hidden Queries"
	chart {
	  name = "Chart Number One"
	  rank = 1
	  type = "timeseries"
	  query {
		hidden              = false
		query_name          = "a"
		display             = "line"
		query_string        = <<EOT
%v
EOT

		hidden_queries = {
			b = "false"
			c = "true"
		}
	  }
	}
  }
`, queryString)

	_ = `	   `

	var dashboard client.UnifiedDashboard
	resourceName := "lightstep_dashboard.test"

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: testAccProviderFactories,
		CheckDestroy:      testGetMetricDashboardDestroy,
		Steps: []resource.TestStep{
			{
				Config: config,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckMetricDashboardExists(resourceName, &dashboard),
					resource.TestCheckResourceAttr(resourceName, "chart.0.query.0.query_string", queryString+"\n"),
					resource.TestCheckResourceAttr(resourceName, "chart.0.query.0.display", "line"),
					resource.TestCheckResourceAttr(resourceName, "chart.0.query.0.hidden", "false"),
					resource.TestCheckResourceAttr(resourceName, "chart.0.query.0.hidden_queries.b", "false"),
					resource.TestCheckResourceAttr(resourceName, "chart.0.query.0.hidden_queries.c", "true"),
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
