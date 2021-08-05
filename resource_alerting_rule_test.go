package main

import (
	"fmt"
	"regexp"

	//"regexp"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/lightstep/terraform-provider-lightstep/lightstep"
)

func TestAccStreamAlertingRule(t *testing.T) {
	var rule lightstep.StreamAlertingRule

	ruleConfig := `
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

resource "lightstep_alerting_rule" "beemo_errors" {
  project_name = ` + fmt.Sprintf("\"%s\"", test_project) + `
  condition_id    = lightstep_stream_condition.beemo_errors.id
  destination_id  = lightstep_slack_destination.slack.id
  update_interval = "1h"
}

resource "lightstep_slack_destination" "slack" {
  project_name = ` + fmt.Sprintf("\"%s\"", test_project) + `
  channel      = "#urgent-care"
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
  condition_name = "Charge errors for BEEMO"
  expression = "err > 0.4"
  evaluation_window_ms = 300000
  stream_id = lightstep_stream.beemo.id
}

resource "lightstep_alerting_rule" "beemo_errors" {
  project_name = ` + fmt.Sprintf("\"%s\"", test_project) + `
  condition_id    = lightstep_stream_condition.beemo_errors.id
  destination_id  = lightstep_slack_destination.slack.id
  # This is being updated!
  update_interval = "2h"
}

resource "lightstep_slack_destination" "slack" {
  project_name = ` + fmt.Sprintf("\"%s\"", test_project) + `
  channel      = "#urgent-care"
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
 expression = "err > 0.4"
 evaluation_window_ms = 300000
 stream_id = lightstep_stream.beemo.id
}

resource "lightstep_alerting_rule" "beemo_errors" {
 project_name = ` + fmt.Sprintf("\"%s\"", test_project) + `
 condition_id    = lightstep_stream_condition.beemo_errors.id
 destination_id  = lightstep_slack_destination.slack.id
 update_interval = "invalid"
}

resource "lightstep_slack_destination" "slack" {
 project_name = ` + fmt.Sprintf("\"%s\"", test_project) + `
 channel      = "#urgent-care"
}
`
	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccAlertingRuleDestroy,
		Steps: []resource.TestStep{
			{
				Config: badExpressionConfig,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAlertingRuleExists("lightstep_alerting_rule.beemo_errors", &rule),
				),
				ExpectError: regexp.MustCompile("got invalid"),
			},
			{
				Config: ruleConfig,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAlertingRuleExists("lightstep_alerting_rule.beemo_errors", &rule),
					resource.TestCheckResourceAttr("lightstep_alerting_rule.beemo_errors", "update_interval", "1h"),
				),
			},
			{
				Config: updatedConfig,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAlertingRuleExists("lightstep_alerting_rule.beemo_errors", &rule),
					// How do you reference the ID of an object created by this test here?
					//resource.TestCheckResourceAttr("lightstep_alerting_rule.beemo_errors", "condition_id", "bar"),
					//resource.TestCheckResourceAttr("lightstep_alerting_rule.beemo_errors", "destination_id", "foo"),
					resource.TestCheckResourceAttr("lightstep_alerting_rule.beemo_errors", "update_interval", "2h"),
				),
			},
		},
	})
}

func TestAccAlertingRuleImport(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:  func() { testAccPreCheck(t) },
		Providers: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: `
resource "lightstep_alerting_rule" "import-cond" {
	project_name = "terraform-provider-tests"
    destination_id = "tvydr9gV"
    condition_id = "KN6SX47x"
	update_interval = "1h"
}
`,
			},
			{
				ResourceName:        "lightstep_alerting_rule.import-cond",
				ImportState:         true,
				ImportStateVerify:   true,
				ImportStateIdPrefix: fmt.Sprintf("%s.", test_project),
			},
		},
	})
}

func testAccCheckAlertingRuleExists(resourceName string, rule *lightstep.StreamAlertingRule) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		tfRule, ok := s.RootModule().Resources[resourceName]
		if !ok {
			return fmt.Errorf("Not found: %s", resourceName)
		}

		if tfRule.Primary.ID == "" {
			return fmt.Errorf("ID is not set")
		}

		client := testAccProvider.Meta().(*lightstep.Client)
		r, err := client.GetAlertingRule(test_project, tfRule.Primary.ID)
		if err != nil {
			return err
		}

		rule = r
		return nil
	}
}

// confirms alerting rules created for test have been destroyed
func testAccAlertingRuleDestroy(s *terraform.State) error {
	conn := testAccProvider.Meta().(*lightstep.Client)

	for _, resource := range s.RootModule().Resources {
		if resource.Type != "alerting_rule" {
			continue
		}

		s, err := conn.GetAlertingRule(test_project, resource.Primary.ID)
		if err == nil {
			if s.ID == resource.Primary.ID {
				return fmt.Errorf("Alerting Rule with ID (%v) still exists.", resource.Primary.ID)
			}
		}
	}
	return nil
}
