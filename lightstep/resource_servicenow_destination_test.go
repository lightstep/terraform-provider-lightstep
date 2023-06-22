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

func TestAccServiceNowDestination(t *testing.T) {
	var destination client.Destination

	missingUrlConfig := `
resource "lightstep_servicenow_destination" "missing_url" {
  project_name = ` + fmt.Sprintf("\"%s\"", testProject) + `
  destination_name = "my-destination"
  auth {
    username = "me"
    password = "pass123"
  }
}
`
	missingAuthConfig := `
resource "lightstep_servicenow_destination" "missing_auth" {
  project_name = ` + fmt.Sprintf("\"%s\"", testProject) + `
  destination_name = "my-destination"
  url = "https://example.com"
}
`

	destinationConfig := `
resource "lightstep_servicenow_destination" "servicenow" {
  project_name = ` + fmt.Sprintf("\"%s\"", testProject) + `
  destination_name = "my-destination"
  url = "https://example.com"
  auth {
    username = "me"
    password = "pass123"
  }
}
`
	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccServiceNowDestinationDestroy,
		Steps: []resource.TestStep{
			{
				Config: missingUrlConfig,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckServiceNowDestinationExists("lightstep_servicenow_destination.missing_url", &destination),
				),
				ExpectError: regexp.MustCompile("The argument \"url\" is required, but no definition was found."),
			},
			{
				Config: missingAuthConfig,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckServiceNowDestinationExists("lightstep_servicenow_destination.missing_auth", &destination),
				),
				ExpectError: regexp.MustCompile("Insufficient auth blocks"),
			},
			{
				Config: destinationConfig,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckWebhookDestinationExists("lightstep_servicenow_destination.servicenow", &destination),
					resource.TestCheckResourceAttr("lightstep_servicenow_destination.servicenow", "destination_name", "my-destination"),
					resource.TestCheckResourceAttr("lightstep_servicenow_destination.servicenow", "url", "https://example.com"),
					resource.TestCheckResourceAttr("lightstep_servicenow_destination.servicenow", "auth.0.username", "me"),
					resource.TestCheckResourceAttr("lightstep_servicenow_destination.servicenow", "auth.0.password", "pass123"),
				),
			},
		},
	})

}

func TestAccServiceNowDestinationImport(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:  func() { testAccPreCheck(t) },
		Providers: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: `
resource "lightstep_servicenow_destination" "servicenow" {
    project_name = ` + fmt.Sprintf("\"%s\"", testProject) + `
	destination_name = "do-not-delete-sn"
	url = "https://example.com"
    auth {
	  username = "me"
	  password = "pass123"
    }
}
`,
			},
			{
				ResourceName:            "lightstep_servicenow_destination.servicenow",
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"auth.0.password"},
				ImportStateIdPrefix:     fmt.Sprintf("%s.", testProject),
			},
		},
	})
}

func testAccCheckServiceNowDestinationExists(resourceName string, destination *client.Destination) resource.TestCheckFunc {
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

func testAccServiceNowDestinationDestroy(s *terraform.State) error {
	conn := testAccProvider.Meta().(*client.Client)
	for _, resource := range s.RootModule().Resources {
		if resource.Type != "lightstep_servicenow_destination" {
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
