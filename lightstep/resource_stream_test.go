package lightstep

import (
	"context"
	"errors"
	"fmt"
	"regexp"
	"strings"
	"testing"

	"github.com/lightstep/terraform-provider-lightstep/client"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
)

func TestAccStream(t *testing.T) {
	var stream client.Stream

	badQuery := `
resource "lightstep_stream" "aggie_errors" {
  project_name = "` + testProject + `"
  stream_name = "Errors (All)"
  query = "error = true"
}
`

	streamConfig := `
resource "lightstep_stream" "aggie_errors" {
  project_name = "` + testProject + `"
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
  project_name = "` + testProject + `"
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
	project_name = "` + testProject + `"
    stream_name = "very important stream to import"
	query = "service IN (\"api\")"
}
`,
			},
			{
				ResourceName:        "lightstep_stream.import-stream",
				ImportState:         true,
				ImportStateVerify:   true,
				ImportStateIdPrefix: fmt.Sprintf("%s.", testProject),
			},
		},
	})
}

func TestAccDeleteStreamWithDependentResources(t *testing.T) {
	makeStreamResource := func(resourceName, streamName string) string {
		return fmt.Sprintf(`
resource "lightstep_stream" "%v" {
	project_name = "%v"
	stream_name  = "%v"
	query        = "service IN (\"web\") AND \"error\" IN (\"true\")"
}
`, resourceName, testProject, streamName)
	}

	streamResourceName1 := "web_errors_1"
	streamName1 := "web errors (1)"
	streamConfig1 := makeStreamResource(streamResourceName1, streamName1)

	streamResourceName2 := "web_errors_2"
	streamName2 := "web errors (2)"
	streamConfig2 := makeStreamResource(streamResourceName2, streamName2)

	// the alert query should map to the stream above
	const alertQuery = `spans count | filter service == "web" && error == true | delta 1h | group_by [], sum`
	alertConfig := fmt.Sprintf(`
resource "lightstep_alert" "web_errors_alert" {
	project_name = "`+testProject+`"
	name = "Span Web Errors alert"

	expression {
		is_multi   = false
		is_no_data = true
		operand    = "above"
		thresholds {
			critical  = 10
			warning = 5
		}
	}

	query {
		hidden       = false
		query_name   = "a"
		display      = "line"
		query_string = <<EOT
			%s
		EOT
	}
}
`, alertQuery)

	var streamCreatedResource1, streamCreatedResource2, streamUpdatedResource client.Stream
	var alertCreatedResource, alertUpdatedResource client.UnifiedCondition
	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccStreamDestroy,
		// each step is akin to running a `terraform apply`
		Steps: []resource.TestStep{
			// create a new stream
			{
				Config: streamConfig1,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckStreamExists("lightstep_stream."+streamResourceName1, &streamCreatedResource1),
				),
			},
			// create a unified alert that implicitly depends on the stream
			{
				Config: streamConfig1 + alertConfig,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckStreamExists("lightstep_stream."+streamResourceName1, &streamUpdatedResource),
					testAccCheckLightstepAlertExists("lightstep_alert.web_errors_alert", &alertCreatedResource),
				),
			},
			// try to delete the stream
			// (the call should succeed and remove the stream from the terraform state, but the stream should still exist)
			{
				Config: alertConfig,
				Check: resource.ComposeTestCheckFunc(
					// verify that terraform still knows about the alert and that it still exists in Lightstep
					testAccCheckLightstepAlertExists("lightstep_alert.web_errors_alert", &alertUpdatedResource),

					// make sure the stream was removed from the terraform state...
					func(state *terraform.State) error {
						err := testAccCheckStreamExists("lightstep_stream."+streamResourceName1, &streamUpdatedResource)(state)
						if err != nil {
							if strings.Contains(err.Error(), "not found: lightstep_stream."+streamResourceName1) {
								return nil
							}
							return errors.New("an unexpected error occurred")
						}
						return errors.New("stream still exists in terraform state")
					},

					// ...but make sure the stream still exists in Lightstep
					func(state *terraform.State) error {
						ctx := context.Background()
						if len(alertUpdatedResource.ID) == 0 {
							// regression check
							return errors.New("unexpected empty alert ID")
						}

						providerClient := testAccProvider.Meta().(*client.Client)
						stream, err := providerClient.GetStream(ctx, testProject, streamCreatedResource1.ID)
						if err != nil {
							return errors.New(fmt.Sprintf("stream not found: %v", err))
						}
						if stream.ID != streamCreatedResource1.ID {
							return errors.New("unexpected stream ID")
						}
						return nil
					},
				),
			},
			// delete the alert that depends on the stream
			{
				Config:  alertConfig,
				Check:   resource.ComposeTestCheckFunc(),
				Destroy: true,
			},
			{
				// create a new stream resource using the original query (it should import/rename the existing stream)
				Config: streamConfig2,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckStreamExists("lightstep_stream."+streamResourceName2, &streamCreatedResource2),
					func(state *terraform.State) error {
						if len(streamCreatedResource1.ID) == 0 {
							return errors.New("unexpected empty stream ID")
						}
						if streamCreatedResource1.ID != streamCreatedResource2.ID {
							return errors.New(fmt.Sprintf("new stream unexpectedly created for same query, streamIDs don't match: '%v' vs '%v'",
								streamCreatedResource1.ID, streamCreatedResource2.ID,
							))
						}
						if streamCreatedResource1.Attributes.Name == streamCreatedResource2.Attributes.Name {
							return errors.New(fmt.Sprintf("new stream unexpectedly has the old name: '%v' vs '%v'",
								streamCreatedResource1.Attributes.Name, streamCreatedResource2.Attributes.Name,
							))
						}
						if streamCreatedResource2.Attributes.Name != streamResourceName2 {
							return errors.New(fmt.Sprintf("new stream has an unexpected name: '%v' vs '%v'",
								streamCreatedResource2.Attributes.Name, streamName2,
							))
						}
						return nil
					},
				),
			},
			// the stream will be deleted automatically as part of test cleanup
		},
	})
}

func TestAccStreamQueryNormalization(t *testing.T) {
	var stream client.Stream

	query1 := `
	resource "lightstep_stream" "query_one" {
	  project_name = "` + testProject + `"
	  stream_name = "Query 1"
	  query = "\"error\" IN (\"true\") AND service IN (\"api\")"
	}
	`
	query1updated := `
	resource "lightstep_stream" "query_one" {
	  project_name = "` + testProject + `"
	  stream_name = "Query One"
	  query = "\"error\" IN (\"true\") AND service IN (\"api\")"
	}
	`
	query1updatedQuery := `
	resource "lightstep_stream" "query_one" {
	  project_name = "` + testProject + `"
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
		str, err := client.GetStream(context.Background(), testProject, tfStream.Primary.ID)
		if err != nil {
			return err
		}

		*stream = *str
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

		s, err := conn.GetStream(context.Background(), testProject, resource.Primary.ID)
		if err == nil {
			if s.ID == resource.Primary.ID {
				return fmt.Errorf("stream with ID (%v) still exists.", resource.Primary.ID)
			}
		}

	}

	return nil
}
