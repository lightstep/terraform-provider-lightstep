package main

import (
	"fmt"
	"regexp"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/terraform"

	"github.com/lightstep/terraform-provider-lightstep/lightstep"
)

func TestAccDestination(t *testing.T) {
	var destination lightstep.Destination

	missingDestination := `
resource "lightstep_destination" "missing_webhook" {
  project_name = ` + fmt.Sprintf("\"%s\"", project) + `
  destination_name = "alert-scraper"
  url = "https://www.alert-scraper.com"
  destination_type="webhook"
  custom_headers = {
    "header_1" = "var_1"
  }
}
`

	destinationConfig := `
resource "lightstep_destination" "webhook" {
  project_name = ` + fmt.Sprintf("\"%s\"", project) + `
  destination_name = "very important webhook"
  url = "https://www.downforeveryoneorjustme.com"
  destination_type="webhook"
}
`
	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccDestinationDestroy,
		Steps: []resource.TestStep{
			{
				Config: missingDestination,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDestinationExists("lightstep_destination.missing_webhook", &destination),
				),
				ExpectError: regexp.MustCompile("config is invalid"),
			},
			{
				Config: destinationConfig,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDestinationExists("lightstep_destination.webhook", &destination),
					resource.TestCheckResourceAttr("lightstep_destination.webhook", "destination_name", "very important webhook"),
					resource.TestCheckResourceAttr("lightstep_destination.webhook", "url", "https://www.downforeveryoneorjustme.com"),
				),
			},
		},
	})

}

func TestAccDestinationImport(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:  func() { testAccPreCheck(t) },
		Providers: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: `
resource "lightstep_destination" "webhook"
	project_name = "terraform-provider-tests"
	destination_name = "do-not-delete"
	url = "https://www.this-is-for-the-integration-tests.com"
	destination_type = "webhook"
`,
			},
			{
				ResourceName:        "lightstep_destination.webhook",
				ImportState:         true,
				ImportStateVerify:   true,
				ImportStateIdPrefix: fmt.Sprintf("%s.", project),
			},
		},
	})
}

func testAccCheckDestinationExists(resourceName string, destination *lightstep.Destination) resource.TestCheckFunc {
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
		d, err := client.GetDestination(project, tfDestination.Primary.ID)
		if err != nil {
			return err
		}

		*destination = d
		return nil
	}
}

func testAccDestinationDestroy(s *terraform.State) error {
	conn := testAccProvider.Meta().(*lightstep.Client)
	for _, resource := range s.RootModule().Resources {
		if resource.Type != "destination" {
			continue
		}

		s, err := conn.GetDestination(project, resource.Primary.ID)
		if err == nil {
			if s.ID == resource.Primary.ID {
				return fmt.Errorf("Destination with ID (%v) still exists.", resource.Primary.ID)
			}
		}

	}
	return nil
}
