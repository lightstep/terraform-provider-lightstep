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

func TestAccPagerdutyDestination(t *testing.T) {
	var destination client.Destination

	missingExpressionConfig := `
resource "lightstep_pagerduty_destination" "missing_pagerduty" {
  project_name = "` + testProject + `"
  destination_name = "missing integration_key"
}
`

	destinationConfig := `
resource "lightstep_pagerduty_destination" "pagerduty" {
  project_name = "` + testProject + `"
  destination_name = "Acceptance Test Destination"
  integration_key = "abc123def456"
}
`
	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccPagerdutyDestinationDestroy,
		Steps: []resource.TestStep{
			{
				Config: missingExpressionConfig,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckPagerdutyDestinationExists("lightstep_pagerduty_destination.missing_pagerduty", &destination),
				),
				ExpectError: regexp.MustCompile("The argument \"integration_key\" is required, but no definition was found."),
			},
			{
				Config: destinationConfig,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckPagerdutyDestinationExists("lightstep_pagerduty_destination.pagerduty", &destination),
					resource.TestCheckResourceAttr("lightstep_pagerduty_destination.pagerduty", "destination_name", "Acceptance Test Destination"),
					resource.TestCheckResourceAttr("lightstep_pagerduty_destination.pagerduty", "integration_key", "abc123def456"),
				),
			},
		},
	})

}

func TestAccPagerdutyDestinationImport(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:  func() { testAccPreCheck(t) },
		Providers: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: `
resource "lightstep_pagerduty_destination" "pagerduty" {
	project_name = "` + testProject + `"
	destination_name = "Terraform LS Destination Acceptance Test Service"
	integration_key = "8e25bec5edc44d05a2acf8238d0246d5"
}
`,
			},
			{
				ResourceName:        "lightstep_pagerduty_destination.pagerduty",
				ImportState:         true,
				ImportStateVerify:   true,
				ImportStateIdPrefix: fmt.Sprintf("%s.", testProject),
			},
		},
	})
}

func testAccCheckPagerdutyDestinationExists(resourceName string, destination *client.Destination) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		// get destination from TF state
		tfDestination, ok := s.RootModule().Resources[resourceName]
		if !ok {
			return fmt.Errorf("not found: %s", resourceName)
		}

		if tfDestination.Primary.ID == "" {
			return fmt.Errorf("id is not set")
		}

		// get destination from LS
		c := testAccProvider.Meta().(*client.Client)
		d, err := c.GetDestination(context.Background(), testProject, tfDestination.Primary.ID)
		if err != nil {
			return err
		}

		destination = d
		return nil
	}
}

func testAccPagerdutyDestinationDestroy(s *terraform.State) error {
	conn := testAccProvider.Meta().(*client.Client)
	for _, resource := range s.RootModule().Resources {
		if resource.Type != "lightstep_pagerduty_destination" {
			continue
		}

		s, err := conn.GetDestination(context.Background(), testProject, resource.Primary.ID)
		if err == nil {
			if s.ID == resource.Primary.ID {
				return fmt.Errorf("destination with ID (%v) still exists.", resource.Primary.ID)
			}
		}

	}
	return nil
}
