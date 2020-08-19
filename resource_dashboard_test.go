package main

import (
	"fmt"
	"regexp"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/terraform"
	"github.com/lightstep/terraform-provider-lightstep/lightstep"
)

func TestAccDashboard(t *testing.T) {
	var dashboard lightstep.Dashboard

	missingStream := `
resource "lightstep_dashboard" "customer_charges" {
  project_name = ` + fmt.Sprintf("\"%s\"", test_project) + `
  dashboard_name = "All Non-BEEMO Charges"
  stream_ids = [lightstep_stream.non_beemo.id]
}
`

	dashboardConfig := `
resource "lightstep_stream" "non_beemo" {
  project_name = "terraform-provider-tests"
  stream_name = "Non-BEEMO charges"
  query = "operation IN (\"api/v1/charge\") AND \"customer_id\" NOT IN (\"BEEMO\")"
}

resource "lightstep_dashboard" "customer_charges" {
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

resource "lightstep_dashboard" "customer_charges" {
 project_name = ` + fmt.Sprintf("\"%s\"", test_project) + `
 dashboard_name = "Customer Charges"
 stream_ids = [lightstep_stream.non_beemo.id]
}
`
	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccDashboardDestroy,
		// each step is akin to running a `terraform apply`
		Steps: []resource.TestStep{
			{
				Config: missingStream,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDashboardExists("lightstep_dashboard.customer_charges", &dashboard),
				),
				ExpectError: regexp.MustCompile("config is invalid"),
			},
			{
				Config: dashboardConfig,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDashboardExists("lightstep_dashboard.customer_charges", &dashboard),
					resource.TestCheckResourceAttr("lightstep_dashboard.customer_charges", "dashboard_name", "All Non-BEEMO Charges"),
				),
			},
			{
				Config: updatedNameDashboard,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDashboardExists("lightstep_dashboard.customer_charges", &dashboard),
					resource.TestCheckResourceAttr("lightstep_dashboard.customer_charges", "dashboard_name", "Customer Charges"),
				),
			},
		},
	})
}

func TestAccDashboardImport(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:  func() { testAccPreCheck(t) },
		Providers: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: `
resource "lightstep_dashboard" "ingress" {
	project_name = "terraform-provider-tests"
 	dashboard_name = "to import"
 	stream_ids = ["VKhpmH1V"]
}
`,
			},
			{
				ResourceName:        "lightstep_dashboard.ingress",
				ImportState:         true,
				ImportStateVerify:   true,
				ImportStateIdPrefix: fmt.Sprintf("%s.", test_project),
			},
		},
	})
}

func testAccCheckDashboardExists(resourceName string, dashboard *lightstep.Dashboard) resource.TestCheckFunc {
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
		client := testAccProvider.Meta().(*lightstep.Client)
		d, err := client.GetDashboard(test_project, tfStream.Primary.ID)
		if err != nil {
			return err
		}

		*dashboard = d
		return nil
	}

}

// confirms that dashboards created during test run have been destroyed
func testAccDashboardDestroy(s *terraform.State) error {
	conn := testAccProvider.Meta().(*lightstep.Client)

	for _, resource := range s.RootModule().Resources {
		if resource.Type != "dashboard" {
			continue
		}

		s, err := conn.GetDashboard(test_project, resource.Primary.ID)
		if err == nil {
			if s.ID == resource.Primary.ID {
				return fmt.Errorf("Dashboard with ID (%v) still exists.", resource.Primary.ID)
			}
		}

	}

	return nil
}
