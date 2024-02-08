package lightstep

import (
	"context"
	"fmt"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/lightstep/terraform-provider-lightstep/client"
	"testing"
)

func TestAccEventQuery(t *testing.T) {
	var eventQuery client.EventQueryAttributes

	eventQueryConfig := `
resource "lightstep_event_query" "terraform" {
  name = "test-name"
  type = "test-type"
  source = "test-source"
  query_string = "logs"
}
`
	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccEventQueryDestroy,
		Steps: []resource.TestStep{
			{
				Config: eventQueryConfig,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckEventQueryExists("lightstep_event_query.terraform", &eventQuery),
					resource.TestCheckResourceAttr("lightstep_event_query.terraform", "name", "test-name"),
					resource.TestCheckResourceAttr("lightstep_event_query.terraform", "type", "test-type"),
					resource.TestCheckResourceAttr("lightstep_event_query.terraform", "source", "test-source"),
					resource.TestCheckResourceAttr("lightstep_event_query.terraform", "query_string", "logs"),
				),
			},
		},
	})

}

func TestAccEventQueryImport(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:  func() { testAccPreCheck(t) },
		Providers: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: `
resource "lightstep_event_query" "imported" {
  name = "test-name-imported"
  type = "test-type-imported"
  source = "test-source-imported"
  query_string = "logs | filter body == imported"
}
`,
			},
			{
				ResourceName:        "lightstep_event_query.imported",
				ImportState:         true,
				ImportStateVerify:   true,
				ImportStateIdPrefix: fmt.Sprintf("%s.", testProject),
			},
		},
	})
}

func testAccCheckEventQueryExists(resourceName string, attrs *client.EventQueryAttributes) resource.TestCheckFunc {
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
		eq, err := client.GetEventQuery(context.Background(), testProject, tfDestination.Primary.ID)
		if err != nil {
			return err
		}

		attrs = eq
		return nil
	}
}

func testAccEventQueryDestroy(s *terraform.State) error {
	conn := testAccProvider.Meta().(*client.Client)
	for _, r := range s.RootModule().Resources {
		if r.Type != "lightstep_event_query" {
			continue
		}

		d, err := conn.GetEventQuery(context.Background(), testProject, r.Primary.ID)
		if err == nil {
			if d.ID == r.Primary.ID {
				return fmt.Errorf("event query with ID (%v) still exists", r.Primary.ID)
			}
		}
	}
	return nil
}
