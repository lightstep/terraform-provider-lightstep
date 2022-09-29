package lightstep

import (
	"context"
	"fmt"
	"os"
	"regexp"
	"testing"

	"github.com/lightstep/terraform-provider-lightstep/client"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
)

const test_project = "terraform-provider-tests"

func TestAccStream(t *testing.T) {
	var stream client.Stream

	badQuery := `
resource "lightstep_stream" "aggie_errors" {
  project_name = ` + fmt.Sprintf("\"%s\"", test_project) + `
  stream_name = "Errors (All)"
  query = "error = true"
}
`

	streamConfig := `
resource "lightstep_stream" "aggie_errors" {
  project_name = ` + fmt.Sprintf("\"%s\"", test_project) + `
  stream_name = "Aggie Errors"
  query = "service IN (\"aggie\") AND \"error\" IN (\"true\")"
  custom_data = [
	  {
		// This name field is special and becomes the key
		"name" = "object1"
		"url" = "https://lightstep.atlassian.net/l/c/M7b0rBsj",
		"key_other" = "value_other",
	  },
  ]
}
`

	updatedNameQuery := `
resource "lightstep_stream" "aggie_errors" {
  project_name = ` + fmt.Sprintf("\"%s\"", test_project) + `
  stream_name = "Errors (All)"
  query = "\"error\" IN (\"true\")"
  custom_data = [
	  {
		// This name field is special and becomes the key
		"name" = "object1"
		"url" = "https://www.lightstep.com",
	  },
  ]
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
					resource.TestCheckResourceAttr("lightstep_stream.aggie_errors", "custom_data.0.name", "object1"),
					resource.TestCheckResourceAttr("lightstep_stream.aggie_errors", "custom_data.0.url", "https://lightstep.atlassian.net/l/c/M7b0rBsj"),
				),
			},
			{
				Config: updatedNameQuery,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckStreamExists("lightstep_stream.aggie_errors", &stream),
					resource.TestCheckResourceAttr("lightstep_stream.aggie_errors", "stream_name", "Errors (All)"),
					resource.TestCheckResourceAttr("lightstep_stream.aggie_errors", "query", "\"error\" IN (\"true\")"),
					resource.TestCheckResourceAttr("lightstep_stream.aggie_errors", "custom_data.0.name", "object1"),
					resource.TestCheckResourceAttr("lightstep_stream.aggie_errors", "custom_data.0.url", "https://www.lightstep.com"),
				),
			},
		},
	})
}

func TestAccStreamImport(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:  func() { testAccPreCheck(t) },
		Providers: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: `
resource "lightstep_stream" "import-stream"{
	project_name = "terraform-provider-tests"
    stream_name = "very important stream to import"
	query = "service IN (\"api\")"
}
`,
			},
			{
				ResourceName:        "lightstep_stream.import-stream",
				ImportState:         true,
				ImportStateVerify:   true,
				ImportStateIdPrefix: fmt.Sprintf("%s.", test_project),
			},
		},
	})
}

func TestAccStreamQueryNormalization(t *testing.T) {
	var stream client.Stream

	query1 := `
	resource "lightstep_stream" "query_one" {
	  project_name = ` + fmt.Sprintf("\"%s\"", test_project) + `
	  stream_name = "Query 1"
	  query = "\"error\" IN (\"true\") AND service IN (\"api\")"
	}
	`
	query1updated := `
	resource "lightstep_stream" "query_one" {
	  project_name = ` + fmt.Sprintf("\"%s\"", test_project) + `
	  stream_name = "Query One"
	  query = "\"error\" IN (\"true\") AND service IN (\"api\")"
	}
	`
	query1updatedQuery := `
	resource "lightstep_stream" "query_one" {
	  project_name = ` + fmt.Sprintf("\"%s\"", test_project) + `
	  stream_name = "Query One"
	  query = "service IN (\"api\") AND \"error\" IN (\"true\")"
	}
	`

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccStreamDestroy,
		// each step is akin to running a `terraform apply`
		Steps: []resource.TestStep{
			{
				Config: query1,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckStreamExists("lightstep_stream.query_one", &stream),
					resource.TestCheckResourceAttr("lightstep_stream.query_one", "stream_name", "Query 1"),
					resource.TestCheckResourceAttr("lightstep_stream.query_one", "query", "\"error\" IN (\"true\") AND service IN (\"api\")"),
				),
			},
			{
				Config: query1updated,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckStreamExists("lightstep_stream.query_one", &stream),
					resource.TestCheckResourceAttr("lightstep_stream.query_one", "stream_name", "Query One"),
					resource.TestCheckResourceAttr("lightstep_stream.query_one", "query", "\"error\" IN (\"true\") AND service IN (\"api\")"),
				),
			},
			{
				Config: query1updatedQuery,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckStreamExists("lightstep_stream.query_one", &stream),
					resource.TestCheckResourceAttr("lightstep_stream.query_one", "stream_name", "Query One"),
					resource.TestCheckResourceAttr("lightstep_stream.query_one", "query", "service IN (\"api\") AND \"error\" IN (\"true\")"),
				),
			},
		},
	})
}

func testAccCheckStreamExists(resourceName string, stream *client.Stream) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		// get stream from TF state
		tfStream, ok := s.RootModule().Resources[resourceName]
		if !ok {
			return fmt.Errorf("not found: %s", resourceName)
		}

		if tfStream.Primary.ID == "" {
			return fmt.Errorf("id is not set")
		}

		// get stream from LS
		client := testAccProvider.Meta().(*client.Client)
		str, err := client.GetStream(context.Background(), test_project, tfStream.Primary.ID)
		if err != nil {
			return err
		}

		stream = str
		return nil
	}

}

// confirms that streams created during test run have been destroyed
func testAccStreamDestroy(s *terraform.State) error {
	conn := testAccProvider.Meta().(*client.Client)

	for _, resource := range s.RootModule().Resources {
		if resource.Type != "stream" {
			continue
		}

		s, err := conn.GetStream(context.Background(), test_project, resource.Primary.ID)
		if err == nil {
			if s.ID == resource.Primary.ID {
				return fmt.Errorf("stream with ID (%v) still exists.", resource.Primary.ID)
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
