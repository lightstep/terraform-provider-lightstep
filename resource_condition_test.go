package main

import (
	"fmt"
	"github.com/hashicorp/terraform-plugin-sdk/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/terraform"
	"github.com/lightstep/terraform-provider-lightstep/lightstep"
	"regexp"
	"testing"
)

func TestAccCondition(t *testing.T) {
	var condition lightstep.Condition

	conditionConfig := `
resource "lightstep_stream" "beemo" {
  project_name = ` + fmt.Sprintf("\"%s\"", project) + `
  stream_name = "BEEMO charges"
  query = "operation IN (\"api/v1/charge\") AND \"customer_id\" IN (\"BEEMO\")"
}

resource "lightstep_condition" "beemo_errors" {
  project_name = ` + fmt.Sprintf("\"%s\"", project) + `
  condition_name = "Charge errors for BEEMO"
  expression = "err > .4"
  evaluation_window_ms = 300000
  stream_id = lightstep_stream.beemo.id
}
`

	updatedConfig := `
resource "lightstep_stream" "beemo" {
  project_name = ` + fmt.Sprintf("\"%s\"", project) + `
  stream_name = "BEEMO charges"
  query = "operation IN (\"api/v1/charge\") AND \"customer_id\" IN (\"BEEMO\")"
}

resource "lightstep_condition" "beemo_errors" {
  project_name = ` + fmt.Sprintf("\"%s\"", project) + `
  condition_name = "Payment Errors for BEEMO"
  expression = "err > .2"
  evaluation_window_ms = 500000
  stream_id = lightstep_stream.beemo.id
}
`
	badExpressionConfig := `
resource "lightstep_stream" "beemo" {
  project_name = ` + fmt.Sprintf("\"%s\"", project) + `
  stream_name = "BEEMO charges"
  query = "operation IN (\"api/v1/charge\") AND \"customer_id\" IN (\"BEEMO\")"
}

resource "lightstep_condition" "beemo_errors" {
  project_name = ` + fmt.Sprintf("\"%s\"", project) + `
  condition_name = "Charge errors for BEEMO"
  expression = "err > 1.4"
  evaluation_window_ms = 300000
  stream_id = lightstep_stream.beemo.id
}
`
	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccConditionDestroy,
		Steps: []resource.TestStep{
			{
				Config: badExpressionConfig,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckConditionExists("lightstep_condition.beemo_errors", &condition),
				),
				ExpectError: regexp.MustCompile("InvalidArgument"),
			},
			{
				Config: conditionConfig,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckConditionExists("lightstep_condition.beemo_errors", &condition),
					resource.TestCheckResourceAttr("lightstep_condition.beemo_errors", "condition_name", "Charge errors for BEEMO"),
					resource.TestCheckResourceAttr("lightstep_condition.beemo_errors", "expression", "err > .4"),
					resource.TestCheckResourceAttr("lightstep_condition.beemo_errors", "evaluation_window_ms", "300000"),
				),
			},
			{
				Config: updatedConfig,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckConditionExists("lightstep_condition.beemo_errors", &condition),
					resource.TestCheckResourceAttr("lightstep_condition.beemo_errors", "condition_name", "Payment Errors for BEEMO"),
					resource.TestCheckResourceAttr("lightstep_condition.beemo_errors", "expression", "err > .2"),
					resource.TestCheckResourceAttr("lightstep_condition.beemo_errors", "evaluation_window_ms", "500000"),
				),
			},
		},
	})
}

func TestAccConditionImport(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:  func() { testAccPreCheck(t) },
		Providers: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: `
resource "lightstep_condition" "import-cond" {
	project_name = "terraform-provider-tests"
	condition_name = "High Ops"
 	expression = "ops > 10000"
 	evaluation_window_ms = 1200000
 	stream_id = "CrwM5g63"
}
`,
			},
			{
				ResourceName:        "lightstep_condition.import-cond",
				ImportState:         true,
				ImportStateVerify:   true,
				ImportStateIdPrefix: fmt.Sprintf("%s.", project),
			},
		},
	})
}

func testAccCheckConditionExists(resourceName string, condition *lightstep.Condition) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		tfCondition, ok := s.RootModule().Resources[resourceName]
		if !ok {
			return fmt.Errorf("Not found: %s", resourceName)
		}

		if tfCondition.Primary.ID == "" {
			return fmt.Errorf("ID is not set")
		}

		client := testAccProvider.Meta().(*lightstep.Client)
		cond, err := client.GetCondition(project, tfCondition.Primary.ID)
		if err != nil {
			return err
		}

		*condition = cond
		return nil
	}
}

// confirms conditions created for test have been destroyed
func testAccConditionDestroy(s *terraform.State) error {
	conn := testAccProvider.Meta().(*lightstep.Client)

	for _, resource := range s.RootModule().Resources {
		if resource.Type != "condition" {
			continue
		}

		s, err := conn.GetCondition(project, resource.Primary.ID)
		if err == nil {
			if s.ID == resource.Primary.ID {
				return fmt.Errorf("Condition with ID (%v) still exists.", resource.Primary.ID)
			}
		}

	}

	return nil
}