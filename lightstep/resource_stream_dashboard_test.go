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

func TestAccStreamDashboard(t *testing.T) {
	var dashboard client.Dashboard

	missingStream := `
resource "lightstep_stream_dashboard" "customer_charges" {
  project_name = ` + fmt.Sprintf("\"%s\"", test_project) + `
  dashboard_name = "All Non-BEEMO Charges"
  stream_ids = ["no-existing"]
}
`

	dashboardConfig := `
resource "lightstep_stream" "non_beemo" {
  project_name = "terraform-provider-tests"
  stream_name = "Non-BEEMO charges"
  query = "operation IN (\"api/v1/charge\") AND \"customer_id\" NOT IN (\"BEEMO\")"
}

resource "lightstep_stream_dashboard" "customer_charges" {
 project_name = ` + fmt.Sprintf("\"%s\"", test_project) + `
 dashboard_name = "All Non-BEEMO Charges"
 stream_ids = [lightstep_stream.non_beemo.id]
}
`
	updatedNameDashboard := `
resource "lightstep_stream" "non_beemo" {
 project_name = ` + fmt.Sprintf("\"%s\"", test_project) + `
  stream_name = "Non-BEEMO charges"
  query = "operation IN (\"api/v1/charge\") AND \"customer_id\" NOT IN (\"BEEMO\")"
}

resource "lightstep_stream_dashboard" "customer_charges" {
 project_name = ` + fmt.Sprintf("\"%s\"", test_project) + `
 dashboard_name = "Customer Charges"
 stream_ids = [lightstep_stream.non_beemo.id]
}
`
	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccStreamDashboardDestroy,
		// each step is akin to running a `terraform apply`
		Steps: []resource.TestStep{
			{
				Config: missingStream,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckStreamDashboardExists("lightstep_stream_dashboard.customer_charges", &dashboard),
				),
				ExpectError: regexp.MustCompile("InvalidArgument"),
			},
			{
				Config: dashboardConfig,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckStreamDashboardExists("lightstep_stream_dashboard.customer_charges", &dashboard),
					resource.TestCheckResourceAttr("lightstep_stream_dashboard.customer_charges", "dashboard_name", "All Non-BEEMO Charges"),
				),
			},
			{
				Config: updatedNameDashboard,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckStreamDashboardExists("lightstep_stream_dashboard.customer_charges", &dashboard),
					resource.TestCheckResourceAttr("lightstep_stream_dashboard.customer_charges", "dashboard_name", "Customer Charges"),
				),
			},
		},
	})
}

func TestAccStreamDashboardImport(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:  func() { testAccPreCheck(t) },
		Providers: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: `
resource "lightstep_stream" "stream_dashboard_source" {
  project_name = ` + fmt.Sprintf("\"%s\"", test_project) + `
  stream_name = "Errors (All)"
  query = "\"error\" IN (\"true\")"
}

resource "lightstep_stream_dashboard" "ingress" {
  project_name = ` + fmt.Sprintf("\"%s\"", test_project) + `
  dashboard_name = "to import"
  stream_ids = [lightstep_stream.stream_dashboard_source.id]
}
`,
			},
			{
				ResourceName:        "lightstep_stream_dashboard.ingress",
				ImportState:         true,
				ImportStateVerify:   true,
				ImportStateIdPrefix: fmt.Sprintf("%s.", test_project),
			},
		},
	})
}

func testAccCheckStreamDashboardExists(resourceName string, dashboard *client.Dashboard) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		// get dashboard from TF state
		tfStream, ok := s.RootModule().Resources[resourceName]
		if !ok {
			return fmt.Errorf("Not found: %s", resourceName)
		}

		if tfStream.Primary.ID == "" {
			return fmt.Errorf("ID is not set")
		}

		// get dashboard from LS
		client := testAccProvider.Meta().(*client.Client)
		d, err := client.GetDashboard(context.Background(), test_project, tfStream.Primary.ID)
		if err != nil {
			return err
		}

		dashboard = d
		return nil
	}

}

// confirms that dashboards created during test run have been destroyed
func testAccStreamDashboardDestroy(s *terraform.State) error {
	conn := testAccProvider.Meta().(*client.Client)

	for _, resource := range s.RootModule().Resources {
		if resource.Type != "dashboard" {
			continue
		}

		s, err := conn.GetDashboard(context.Background(), test_project, resource.Primary.ID)
		if err == nil {
			if s.ID == resource.Primary.ID {
				return fmt.Errorf("Dashboard with ID (%v) still exists.", resource.Primary.ID)
			}
		}

	}

	return nil
}
