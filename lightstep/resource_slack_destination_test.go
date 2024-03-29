package lightstep

import (
	"context"
	"fmt"
	"testing"

	"github.com/lightstep/terraform-provider-lightstep/client"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
)

func TestAccSlackDestination(t *testing.T) {
	var destination client.Destination

	destinationConfig := `
resource "lightstep_slack_destination" "slack" {
  project_name = "` + testProject + `"
  channel = "#emergency-room"
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
 project_name = "` + testProject + `"
 channel = "#terraform-provider-acceptance-tests"
}
`,
			},
			{
				ResourceName:        "lightstep_slack_destination.imported",
				ImportState:         true,
				ImportStateVerify:   true,
				ImportStateIdPrefix: fmt.Sprintf("%s.", testProject),
			},
		},
	})
}

func testAccCheckSlackDestinationExists(resourceName string, destination *client.Destination) resource.TestCheckFunc {
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
		client := testAccProvider.Meta().(*client.Client)
		d, err := client.GetDestination(context.Background(), testProject, tfDestination.Primary.ID)
		if err != nil {
			return err
		}

		destination = d
		return nil
	}
}

func testAccSlackDestinationDestroy(s *terraform.State) error {
	conn := testAccProvider.Meta().(*client.Client)
	for _, r := range s.RootModule().Resources {
		if r.Type != "lightstep_slack_destination" {
			continue
		}

		d, err := conn.GetDestination(context.Background(), testProject, r.Primary.ID)
		if err == nil {
			if d.ID == r.Primary.ID {
				return fmt.Errorf("destination with ID (%v) still exists.", r.Primary.ID)
			}
		}
	}
	return nil
}
