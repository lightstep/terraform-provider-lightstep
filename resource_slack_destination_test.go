package main

import (
	"fmt"
	"regexp"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/terraform"

	"github.com/lightstep/terraform-provider-lightstep/lightstep"
)

func TestAccSlackDestination(t *testing.T) {
	var destination lightstep.Destination

	//	destinationConfig := `
	//resource "lightstep_slack_destination" "slack" {
	// project_name = "terraform-provider-tests"
	// channel = "#emergency-room"
	//}
	//`

	destinationConfig := `
resource "lightstep_slack_destination" "slack" {
	project_name = "terraform-provider-tests"
	channel = "#emergency-room"
}
`
	missingExpressionConfig := `
	resource "lightstep_slack_destination" "missing_channel" {
	project_name = "terraform-provider-tests"
	}
	`

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccSlackDestinationDestroy,
		Steps: []resource.TestStep{
			{
				Config: destinationConfig,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckSlackDestinationExists("lightstep_slack_destination.slack", &destination),
					resource.TestCheckResourceAttr("lightstep_slack_destination.slack", "channel", "#emergency-room"),
				),
			},
			{
				Config: missingExpressionConfig,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckSlackDestinationExists("lightstep_slack_destination.missing_channel", &destination),
				),
				ExpectError: regexp.MustCompile("Missing required argument: The argument \"channel\" is required, but no definition was found."),
			},
		},
	})

}

func TestAccSlackDestinationImport(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:  func() { testAccPreCheck(t) },
		Providers: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: `
resource "lightstep_slack_destination" "imported" {
 project_name = "terraform-provider-tests"
 channel = "#terraform-provider-acceptance-tests"
}
`,
			},
			{
				ResourceName:        "lightstep_slack_destination.imported",
				ImportState:         true,
				ImportStateVerify:   true,
				ImportStateIdPrefix: fmt.Sprintf("%s.", test_project),
			},
		},
	})
}

func testAccCheckSlackDestinationExists(resourceName string, destination *lightstep.Destination) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		// get destination from TF state
		tfDestination, ok := s.RootModule().Resources[resourceName]
		if !ok {
			return fmt.Errorf("Not found: %s", resourceName)
		}

		if tfDestination.Primary.ID == "" {
			return fmt.Errorf("ID is not set")
		}

		// get destination from LS
		client := testAccProvider.Meta().(*lightstep.Client)
		d, err := client.GetDestination(test_project, tfDestination.Primary.ID)
		if err != nil {
			return err
		}

		*destination = d
		return nil
	}
}

// WHY ISN't DESTROY WORKING

func testAccSlackDestinationDestroy(s *terraform.State) error {
	conn := testAccProvider.Meta().(*lightstep.Client)
	for _, r := range s.RootModule().Resources {
		if r.Type != "lightstep_slack_destination" {
			continue
		}

		_, err := conn.GetDestination(test_project, r.Primary.ID)
		if err == nil {
			return fmt.Errorf("Destination with ID (%v) still exists.", r.Primary.ID)

		}

	}
	return nil
}
