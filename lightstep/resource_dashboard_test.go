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

func TestLegacyImplicitGroup(t *testing.T) {
	var dashboard client.UnifiedDashboard

	baseConfig := `
resource "lightstep_dashboard" "test_implicit_group" {
  project_name   = "terraform-provider-tests"
  dashboard_name = "test legacy implicit groups"

  # we declare this implicit group explicitly; it is thus _not_ a legacy implicit group
  group {
    rank            = 0
    title           = ""
    visibility_type = "implicit"

    chart {
      name   = "Kool Khart"
      type   = "timeseries"
      rank   = 0
      x_pos  = 0
      y_pos  = 0
      width  = 0
      height = 0

      query {
        query_name   = "a"
        display      = "line"
        hidden       = false
        query_string = "metric cpu.utilization | delta | group_by[], sum"
      }
    }
  }
}
`

	resourceName := "lightstep_dashboard.test_implicit_group"

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testGetMetricDashboardDestroy,
		Steps: []resource.TestStep{
			{
				Config: baseConfig,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckMetricDashboardExists(resourceName, &dashboard),
					resource.TestCheckResourceAttr(resourceName, "dashboard_name", "test legacy implicit groups"),
					resource.TestCheckResourceAttr(resourceName, "chart.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "group.#", "1"),
				),
			},
		},
	})
}

func TestGroupChartsAreComputed(t *testing.T) {
	var dashboard client.UnifiedDashboard

	baseConfig := `
resource "lightstep_dashboard" "test_dash" {
  project_name   = "terraform-provider-tests"
  dashboard_name = "test dash"

  label {
    key   = "type"
    value = "service"
  }

  group {
    rank            = 0
    title           = ""
    visibility_type = "implicit"
    chart {
      name   = "chart 1"
      type   = "timeseries"
      rank   = 1
      x_pos  = 0
      y_pos  = 0
      width  = 16
      height = 10

      query {
        query_name   = "a"
        display      = "area"
        hidden       = false
        query_string = "metric foo | rate | group_by [], sum"
      }
    }
	chart {
      name   = "chart 2"
      type   = "timeseries"
      rank   = 3
      x_pos  = 0
      y_pos  = 0
      width  = 0
      height = 0
      query {
        query_name     = "(b-a)"
        display        = "line"
        hidden         = false
        hidden_queries = { a = "true", b = "true" }
        query_string   = "with\n  a = metric cpu.utilization | filter (kube_app == \"iOS\") | latest | group_by [], mean;\n  b = metric cpu.utilization | filter (kube_app == \"android\") | latest | group_by [], mean;\njoin ((b - a)), a=0, b=0"
      }
    }
  }
}
`

	updatedConfig := `
resource "lightstep_dashboard" "test_dash" {
  project_name   = "terraform-provider-tests"
  dashboard_name = "test dash"

  label {
    key   = "type"
    value = "service"
  }

  group {
    rank            = 0
    title           = ""
    visibility_type = "implicit"
    chart {
      name   = "chart 1!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!"
      type   = "timeseries"
      rank   = 1
      x_pos  = 0
      y_pos  = 0
      width  = 16
      height = 10

      query {
        query_name   = "a"
        display      = "area"
        hidden       = false
        query_string = "metric foo | rate | group_by [], sum"
      }
    }
	chart {
      name   = "chart 2"
      type   = "timeseries"
      rank   = 3
      x_pos  = 0
      y_pos  = 0
      width  = 0
      height = 0
      query {
        query_name     = "(b-a)"
        display        = "line"
        hidden         = false
        hidden_queries = { a = "true", b = "true" }
        query_string   = "with\n  a = metric cpu.utilization | filter (kube_app == \"iOS\") | latest | group_by [], mean;\n  b = metric cpu.utilization | filter (kube_app == \"android\") | latest | group_by [], mean;\njoin ((b - a)), a=0, b=0"
      }
    }
  }
}
`

	resourceName := "lightstep_dashboard.test_dash"

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testGetMetricDashboardDestroy,
		Steps: []resource.TestStep{
			{
				Config: baseConfig,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckMetricDashboardExists(resourceName, &dashboard),
					resource.TestCheckResourceAttr(resourceName, "dashboard_name", "test dash"),
					resource.TestCheckResourceAttr(resourceName, "chart.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "group.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "group.0.chart.#", "2"),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "group.0.chart.*",
						map[string]string{
							"name": "chart 1",
						},
					),
				),
			},
			{
				Config: updatedConfig,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckMetricDashboardExists(resourceName, &dashboard),
					resource.TestCheckResourceAttr(resourceName, "dashboard_name", "test dash"),
					resource.TestCheckResourceAttr(resourceName, "chart.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "group.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "group.0.chart.#", "2"),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "group.0.chart.*",
						map[string]string{
							"name": "chart 1!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!",
						},
					),
				),
			},
		},
	})
}

func makeDisplayTypeConfig(displayType, displayTypeOptions string) string {
	return fmt.Sprintf(`
resource "lightstep_dashboard" "test_display_type_options" {
project_name   = "terraform-provider-tests"
dashboard_name = "test display_type_options"

group {
rank            = 0
title           = ""
visibility_type = "implicit"

chart {
	name   = "Chart #123"
	type   = "timeseries"
	rank   = 0
	x_pos  = 4
	y_pos  = 0
	width  = 4
	height = 4

	query {
	  query_name   = "a"
	  display      = "%v"
	  %v
	  hidden       = false
	  query_string = "metric cpu.utilization | delta | group_by[], sum"
	}
  }
}
}
`, displayType, displayTypeOptions)
}

func TestDisplayTypeOptionsError(t *testing.T) {
	var dashboard client.UnifiedDashboard

	resourceName := "lightstep_dashboard.test_display_type_options"

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testGetMetricDashboardDestroy,
		Steps: []resource.TestStep{
			{
				Config: makeDisplayTypeConfig("table", strings.TrimSpace(`
	display_type_options {
			positive_data_only = true
	}
`)),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckMetricDashboardExists(resourceName, &dashboard),
				),
				ExpectError: regexp.MustCompile("Unsupported argument"),
			},
		},
	})
}

func TestDisplayTypeOptions(t *testing.T) {
	var dashboard client.UnifiedDashboard

	resourceName := "lightstep_dashboard.test_display_type_options"

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testGetMetricDashboardDestroy,
		Steps: []resource.TestStep{
			{
				Config: makeDisplayTypeConfig("line", ""),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckMetricDashboardExists(resourceName, &dashboard),
					resource.TestCheckResourceAttr(resourceName, "dashboard_name", "test display_type_options"),
					resource.TestCheckResourceAttr(resourceName, "group.0.chart.0.query.0.display", "line"),
				),
			},
			{
				Config: makeDisplayTypeConfig("ordered_list", strings.TrimSpace(`
	display_type_options {
			sort_direction = "asc"
	}
`)),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckMetricDashboardExists(resourceName, &dashboard),
					resource.TestCheckResourceAttr(resourceName, "dashboard_name", "test display_type_options"),
					resource.TestCheckResourceAttr(resourceName, "group.0.chart.0.query.0.display", "ordered_list"),
					resource.TestCheckResourceAttr(resourceName, "group.0.chart.0.query.0.display_type_options.0.sort_direction", "asc"),
				),
			},
			{
				Config: makeDisplayTypeConfig("table", strings.TrimSpace(`
	display_type_options {
			sort_direction = "desc"
			sort_by = "value"
	}
`)),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckMetricDashboardExists(resourceName, &dashboard),
					resource.TestCheckResourceAttr(resourceName, "dashboard_name", "test display_type_options"),
					resource.TestCheckResourceAttr(resourceName, "group.0.chart.0.query.0.display", "table"),
					resource.TestCheckResourceAttr(resourceName, "group.0.chart.0.query.0.display_type_options.0.sort_direction", "desc"),
					resource.TestCheckResourceAttr(resourceName, "group.0.chart.0.query.0.display_type_options.0.sort_by", "value"),
				),
			},
		},
	})
}
