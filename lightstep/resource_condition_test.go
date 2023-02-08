package lightstep

import (
	"regexp"
	"testing"

	"github.com/lightstep/terraform-provider-lightstep/client"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

func TestAccCondition(t *testing.T) {
	var condition client.UnifiedCondition

	badCondition := `
resource "lightstep_condition" "errors" {
  project_name = "terraform-provider-tests"
  name = "Too many requests"
  expression {
	  is_multi   = true
	  is_no_data = true
      operand  = "above"
	
	  thresholds {
		critical  = 10
		warning = 5
	  }
  }
}
`

	conditionConfig := `
resource "lightstep_slack_destination" "slack" {
  project_name = "terraform-provider-tests"
  channel = "#emergency-room"
}

resource "lightstep_condition" "test" {
  project_name = "terraform-provider-tests"
  name = "Too many requests"
  description = "A link to a playbook"

  expression {
	  is_multi   = true
	  is_no_data = true
      operand  = "above"
	  thresholds {
		critical  = 10
		warning = 5
	  }
  }

  query {
    query_name          = "a"
    hidden              = false
    display 			= "line"
	query_string        = "metric requests | rate 1h, 30s | filter \"project_name\" == \"catlab\" && \"service\" != \"android\" | group_by[\"method\"], mean | reduce 30s, min"
  }

  alerting_rule {
    id          = lightstep_slack_destination.slack.id
    update_interval = "1h"

    include_filters = [
      {
        key   = "project_name"
        value = "catlab"
      }
    ]

	filters = [
		{
		  key   = "service_name"
		  value = "frontend"
		  operand = "contains"
		}
	  ]
  }
}
`

	updatedConditionConfig := `
resource "lightstep_slack_destination" "slack" {
  project_name = "terraform-provider-tests"
  channel = "#emergency-room"
}

resource "lightstep_condition" "test" {
  project_name = "terraform-provider-tests"
  name = "updated"
  description = "A link to a fresh playbook"

  expression {
	  is_multi   = true
	  is_no_data = false
      operand  = "above"
	  thresholds {
		critical  = 10
		warning = 5
	  }
  }

  query {
    query_name          = "a"
    hidden              = false
    display 			= "line"
	query_string        = "metric requests | rate 1h, 30s | filter \"project_name\" == \"catlab\" && \"service\" != \"iOS\" | group_by[\"method\"], mean | reduce 30s, min"
  }

  alerting_rule {
    id          = lightstep_slack_destination.slack.id
    update_interval = "1h"

    include_filters = [
      {
        key   = "project_name"
        value = "catlab"
      }
    ]
  }
}
`

	resourceName := "lightstep_condition.test"
	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccMetricConditionDestroy,
		Steps: []resource.TestStep{
			{
				Config: badCondition,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckMetricConditionExists(resourceName, &condition),
				),
				ExpectError: regexp.MustCompile("(Missing required argument|Insufficient query blocks)"),
			},
			{
				Config: conditionConfig,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckMetricConditionExists(resourceName, &condition),
					resource.TestCheckResourceAttr(resourceName, "name", "Too many requests"),
					resource.TestCheckResourceAttr(resourceName, "description", "A link to a playbook"),
					resource.TestCheckResourceAttr(resourceName, "query.0.query_string", "metric requests | rate 1h, 30s | filter \"project_name\" == \"catlab\" && \"service\" != \"android\" | group_by[\"method\"], mean | reduce 30s, min"),
					resource.TestCheckResourceAttr(resourceName, "expression.0.is_no_data", "true"),
				),
			},
			{
				Config: updatedConditionConfig,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckMetricConditionExists(resourceName, &condition),
					resource.TestCheckResourceAttr(resourceName, "name", "updated"),
					resource.TestCheckResourceAttr(resourceName, "description", "A link to a fresh playbook"),
					resource.TestCheckResourceAttr(resourceName, "query.0.query_string", "metric requests | rate 1h, 30s | filter \"project_name\" == \"catlab\" && \"service\" != \"iOS\" | group_by[\"method\"], mean | reduce 30s, min"),
					resource.TestCheckResourceAttr(resourceName, "expression.0.is_no_data", "false"),
				),
			},
		},
	})
}
