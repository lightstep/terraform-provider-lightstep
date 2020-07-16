package main

import (
	"fmt"
	"github.com/hashicorp/terraform-plugin-sdk/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/terraform"
	"github.com/lightstep/terraform-provider-lightstep/lightstep"
	"os"
	"regexp"
	"testing"
)

const project = "terraform-provider-tests"

func TestAccStream(t *testing.T) {
	var stream lightstep.Stream

	badQuery := `
resource "lightstep_stream" "aggie_errors" {
  project_name = ` + fmt.Sprintf("\"%s\"", project) + `
  stream_name = "Errors (All)"
  query = "error = true"
}
`

	streamConfig := `
resource "lightstep_stream" "aggie_errors" {
  project_name = ` + fmt.Sprintf("\"%s\"", project) + `
  stream_name = "Aggie Errors"
  query = "service IN (\"aggie\") AND \"error\" IN (\"true\")"
}
`

	updatedNameQuery := `
resource "lightstep_stream" "aggie_errors" {
  project_name = ` + fmt.Sprintf("\"%s\"", project) + `
  stream_name = "Errors (All)"
  query = "\"error\" IN (\"true\")"
}
`
	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccStreamDestroy,
		// each step is akin to running a `terraform apply`
		Steps: []resource.TestStep{
			{
				Config: badQuery,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckStreamExists("lightstep_stream.aggie_errors", &stream),
				),
				ExpectError: regexp.MustCompile("InvalidArgument"),
			},
			{
				Config: streamConfig,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckStreamExists("lightstep_stream.aggie_errors", &stream),
					resource.TestCheckResourceAttr("lightstep_stream.aggie_errors", "stream_name", "Aggie Errors"),
					resource.TestCheckResourceAttr("lightstep_stream.aggie_errors", "query", "service IN (\"aggie\") AND \"error\" IN (\"true\")"),
				),
			},
			{
				Config: updatedNameQuery,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckStreamExists("lightstep_stream.aggie_errors", &stream),
					resource.TestCheckResourceAttr("lightstep_stream.aggie_errors", "stream_name", "Errors (All)"),
					resource.TestCheckResourceAttr("lightstep_stream.aggie_errors", "query", "\"error\" IN (\"true\")"),
				),
			},
		},
	})
}

func testAccCheckStreamExists(resourceName string, stream *lightstep.Stream) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		// get stream from TF state
		tfStream, ok := s.RootModule().Resources[resourceName]
		if !ok {
			return fmt.Errorf("Not found: %s", resourceName)
		}

		if tfStream.Primary.ID == "" {
			return fmt.Errorf("ID is not set")
		}

		// get stream from LS
		client := testAccProvider.Meta().(*lightstep.Client)
		str, err := client.GetStream(project, tfStream.Primary.ID)
		if err != nil {
			return err
		}

		*stream = str
		return nil
	}

}

// confirms that streams created during test run have been destroyed
func testAccStreamDestroy(s *terraform.State) error {
	conn := testAccProvider.Meta().(*lightstep.Client)

	for _, resource := range s.RootModule().Resources {
		if resource.Type != "stream" {
			continue
		}

		s, err := conn.GetStream(project, resource.Primary.ID)
		if err == nil {
			if s.ID == resource.Primary.ID {
				return fmt.Errorf("Stream with ID (%v) still exists.", resource.Primary.ID)
			}
		}

	}

	return nil
}

func testAccPreCheck(t *testing.T) {
	if v := os.Getenv("LIGHTSTEP_API_KEY"); v == "" {
		t.Fatal("LIGHTSTEP_API_KEY must be set.")
	}
	if v := os.Getenv("LIGHTSTEP_ORG"); v == "" {
		t.Fatal("LIGHTSTEP_ORG must be set")
	}
}
