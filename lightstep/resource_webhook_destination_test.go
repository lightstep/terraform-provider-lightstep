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

func TestAccWebhookDestination(t *testing.T) {
	var destination client.Destination

	missingExpressionConfig := `
resource "lightstep_webhook_destination" "missing_webhook" {
  project_name = "` + testProject + `"
  destination_name = "alert-scraper"
}
`

	destinationConfig := `
resource "lightstep_webhook_destination" "webhook" {
  project_name = "` + testProject + `"
  destination_name = "very important webhook"
  url = "https://www.downforeveryoneorjustme.com"
  custom_headers = {
  	"header_1" = "value_1"
    "header_2" = "value_2"
  }
}
`
	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccWebhookDestinationDestroy,
		Steps: []resource.TestStep{
			{
				Config: missingExpressionConfig,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckWebhookDestinationExists("lightstep_webhook_destination.missing_webhook", &destination),
				),
				ExpectError: regexp.MustCompile("The argument \"url\" is required, but no definition was found."),
			},
			{
				Config: destinationConfig,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckWebhookDestinationExists("lightstep_webhook_destination.webhook", &destination),
					resource.TestCheckResourceAttr("lightstep_webhook_destination.webhook", "destination_name", "very important webhook"),
					resource.TestCheckResourceAttr("lightstep_webhook_destination.webhook", "url", "https://www.downforeveryoneorjustme.com"),
				),
			},
		},
	})

}

func TestAccWebhookDestinationImport(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:  func() { testAccPreCheck(t) },
		Providers: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: `
resource "lightstep_webhook_destination" "webhook" {
	project_name = "` + testProject + `"
	destination_name = "do-not-delete"
	url = "https://www.this-is-for-the-integration-tests.com"
    custom_headers = {
	  "allow-all" = "forever"
    }
}
`,
			},
			{
				ResourceName:        "lightstep_webhook_destination.webhook",
				ImportState:         true,
				ImportStateVerify:   true,
				ImportStateIdPrefix: fmt.Sprintf("%s.", testProject),
			},
		},
	})
}

func testAccCheckWebhookDestinationExists(resourceName string, destination *client.Destination) resource.TestCheckFunc {
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

func testAccWebhookDestinationDestroy(s *terraform.State) error {
	conn := testAccProvider.Meta().(*client.Client)
	for _, resource := range s.RootModule().Resources {
		if resource.Type != "lightstep_webhook_destination" {
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
