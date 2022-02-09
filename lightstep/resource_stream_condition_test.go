package lightstep

import (
	"context"
	"fmt"
	"github.com/lightstep/terraform-provider-lightstep/client"
	"regexp"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
)

func TestAccStreamCondition(t *testing.T) {
	var condition client.StreamCondition

	conditionConfig := `
resource "lightstep_stream" "beemo" {
  project_name = ` + fmt.Sprintf("\"%s\"", test_project) + `
  stream_name = "BEEMO charges"
  query = "operation IN (\"api/v1/charge\") AND \"customer_id\" IN (\"BEEMO\")"
}

resource "lightstep_stream_condition" "beemo_errors" {
  project_name = ` + fmt.Sprintf("\"%s\"", test_project) + `
  condition_name = "Charge errors for BEEMO"
  expression = "err > 0.4"
  evaluation_window_ms = 300000
  stream_id = lightstep_stream.beemo.id
}
`

	updatedConfig := `
resource "lightstep_stream" "beemo" {
  project_name = ` + fmt.Sprintf("\"%s\"", test_project) + `
  stream_name = "BEEMO charges"
  query = "operation IN (\"api/v1/charge\") AND \"customer_id\" IN (\"BEEMO\")"
}

resource "lightstep_stream_condition" "beemo_errors" {
  project_name = ` + fmt.Sprintf("\"%s\"", test_project) + `
  condition_name = "Payment Errors for BEEMO"
  expression = "err > 0.2"
  evaluation_window_ms = 500000
  stream_id = lightstep_stream.beemo.id
}
`
	badExpressionConfig := `
resource "lightstep_stream" "beemo" {
  project_name = ` + fmt.Sprintf("\"%s\"", test_project) + `
  stream_name = "BEEMO charges"
  query = "operation IN (\"api/v1/charge\") AND \"customer_id\" IN (\"BEEMO\")"
}

resource "lightstep_stream_condition" "beemo_errors" {
  project_name = ` + fmt.Sprintf("\"%s\"", test_project) + `
  condition_name = "Charge errors for BEEMO"
  expression = "err > 1.4"
  evaluation_window_ms = 300000
  stream_id = lightstep_stream.beemo.id
}
`
	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccStreamConditionDestroy,
		Steps: []resource.TestStep{
			{
				Config: badExpressionConfig,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckStreamConditionExists("lightstep_stream_condition.beemo_errors", &condition),
				),
				ExpectError: regexp.MustCompile("InvalidArgument"),
			},
			{
				Config: conditionConfig,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckStreamConditionExists("lightstep_stream_condition.beemo_errors", &condition),
					resource.TestCheckResourceAttr("lightstep_stream_condition.beemo_errors", "condition_name", "Charge errors for BEEMO"),
					// the lightstep API returns the threshold as 0.4 and not .4
					resource.TestCheckResourceAttr("lightstep_stream_condition.beemo_errors", "expression", "err > 0.4"),
					resource.TestCheckResourceAttr("lightstep_stream_condition.beemo_errors", "evaluation_window_ms", "300000"),
				),
			},
			{
				Config: updatedConfig,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckStreamConditionExists("lightstep_stream_condition.beemo_errors", &condition),
					resource.TestCheckResourceAttr("lightstep_stream_condition.beemo_errors", "condition_name", "Payment Errors for BEEMO"),
					// the lightstep API returns the threshold as 0.2 and not .2
					resource.TestCheckResourceAttr("lightstep_stream_condition.beemo_errors", "expression", "err > 0.2"),
					resource.TestCheckResourceAttr("lightstep_stream_condition.beemo_errors", "evaluation_window_ms", "500000"),
				),
			},
		},
	})
}

func TestAccStreamConditionImport(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:  func() { testAccPreCheck(t) },
		Providers: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: `
resource "lightstep_stream" "import_stream_condition_stream" {
  project_name = ` + fmt.Sprintf("\"%s\"", test_project) + `
  stream_name = "Stream Condition Import charges"
  query = "operation IN (\"api/v1/stream\")"
}

resource "lightstep_stream_condition" "import-cond" {
  project_name = ` + fmt.Sprintf("\"%s\"", test_project) + `
  condition_name = "High Ops"
  expression = "ops > 10000"
  evaluation_window_ms = 1200000
  stream_id = lightstep_stream.import_stream_condition_stream.id
}
`,
			},
			{
				ResourceName:        "lightstep_stream_condition.import-cond",
				ImportState:         true,
				ImportStateVerify:   true,
				ImportStateIdPrefix: fmt.Sprintf("%s.", test_project),
			},
		},
	})
}

func testAccCheckStreamConditionExists(resourceName string, condition *client.StreamCondition) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		tfCondition, ok := s.RootModule().Resources[resourceName]
		if !ok {
			return fmt.Errorf("Not found: %s", resourceName)
		}

		if tfCondition.Primary.ID == "" {
			return fmt.Errorf("ID is not set")
		}

		c := testAccProvider.Meta().(*client.Client)
		cond, err := c.GetStreamCondition(context.Background(), test_project, tfCondition.Primary.ID)
		if err != nil {
			return err
		}

		condition = cond
		return nil
	}
}

// confirms conditions created for test have been destroyed
func testAccStreamConditionDestroy(s *terraform.State) error {
	conn := testAccProvider.Meta().(*client.Client)

	for _, resource := range s.RootModule().Resources {
		if resource.Type != "condition" {
			continue
		}

		s, err := conn.GetStreamCondition(context.Background(), test_project, resource.Primary.ID)
		if err == nil {
			if s.ID == resource.Primary.ID {
				return fmt.Errorf("Condition with ID (%v) still exists.", resource.Primary.ID)
			}
		}
	}
	return nil
}
