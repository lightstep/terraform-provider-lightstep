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

func TestAccAlert(t *testing.T) {
	var condition client.UnifiedCondition

	badAlertMissingQueryAndCompositeFields := `
resource "lightstep_alert" "errors" {
  project_name = "` + testProject + `"
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
  project_name = "` + testProject + `"
  channel = "#emergency-room"
}

resource "lightstep_alert" "test" {
  project_name = "` + testProject + `"
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
	query_string        = "metric requests | rate 1h, 30s | filter \"project_name\" == \"catlab\" && \"service\" != \"android\" | group_by[\"method\"], mean | reduce 30s, min"
  }

  alerting_rule {
    id          = lightstep_slack_destination.slack.id
    update_interval = "1h"
  }
}
`

	updatedConditionConfig := `
resource "lightstep_slack_destination" "slack" {
  project_name = "` + testProject + `"
  channel = "#emergency-room"
}

resource "lightstep_alert" "test" {
  project_name = "` + testProject + `"
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
	query_string        = "metric requests | rate 1h, 30s | filter \"project_name\" == \"catlab\" && \"service\" != \"iOS\" | group_by[\"method\"], mean | reduce 30s, min"
  }

  alerting_rule {
    id          = lightstep_slack_destination.slack.id
    update_interval = "1h"
  }
}
`

	resourceName := "lightstep_alert.test"
	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccMetricConditionDestroy,
		Steps: []resource.TestStep{
			{
				Config: badAlertMissingQueryAndCompositeFields,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckLightstepAlertExists(resourceName, &condition),
				),
				ExpectError: regexp.MustCompile("at least one query is required"),
			},
			{
				Config: conditionConfig,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckLightstepAlertExists(resourceName, &condition),
					resource.TestCheckResourceAttr(resourceName, "name", "Too many requests"),
					resource.TestCheckResourceAttr(resourceName, "description", "A link to a playbook"),
					resource.TestCheckResourceAttr(resourceName, "query.0.query_string", "metric requests | rate 1h, 30s | filter \"project_name\" == \"catlab\" && \"service\" != \"android\" | group_by[\"method\"], mean | reduce 30s, min"),
					resource.TestCheckResourceAttr(resourceName, "expression.0.is_no_data", "true"),
				),
			},
			{
				Config: updatedConditionConfig,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckLightstepAlertExists(resourceName, &condition),
					resource.TestCheckResourceAttr(resourceName, "name", "updated"),
					resource.TestCheckResourceAttr(resourceName, "description", "A link to a fresh playbook"),
					resource.TestCheckResourceAttr(resourceName, "query.0.query_string", "metric requests | rate 1h, 30s | filter \"project_name\" == \"catlab\" && \"service\" != \"iOS\" | group_by[\"method\"], mean | reduce 30s, min"),
					resource.TestCheckResourceAttr(resourceName, "expression.0.is_no_data", "false"),
				),
			},
		},
	})
}

func TestAccAlert2(t *testing.T) {
	var condition client.UnifiedCondition

	uqlQuery := `metric requests | filter ((service != "android") && (project_name == "catlab")) | rate 1h, 1h | group_by ["method"], mean | reduce 5m, min`

	badCondition := `
resource "lightstep_metric_condition" "errors" {
  project_name = "` + testProject + `"
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

	conditionConfig := fmt.Sprintf(`
resource "lightstep_slack_destination" "slack" {
  project_name = "`+testProject+`"
  channel = "#emergency-room"
}

resource "lightstep_alert" "test" {
  project_name = "`+testProject+`"
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
    query_name      = "a"
    hidden          = false
    display 		= "line"
	query_string 	= <<EOT
%s
EOT
  }

  alerting_rule {
    id          = lightstep_slack_destination.slack.id
    update_interval = "1h"
  }
}
`, uqlQuery)

	updatedConditionConfig := fmt.Sprintf(`
resource "lightstep_slack_destination" "slack" {
  project_name = "`+testProject+`"
  channel = "#emergency-room"
}

resource "lightstep_alert" "test" {
  project_name = "`+testProject+`"
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
	display             = "line"
	query_string 	= <<EOT
%s
EOT
	
  }

  alerting_rule {
    id          = lightstep_slack_destination.slack.id
    update_interval = "1h"
  }
}
`, uqlQuery)

	resourceName := "lightstep_alert.test"
	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccMetricConditionDestroy,
		Steps: []resource.TestStep{
			{
				Config: badCondition,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckLightstepAlertExists(resourceName, &condition),
				),
				ExpectError: regexp.MustCompile("(Missing required argument|Insufficient metric_query blocks)"),
			},
			{
				Config: conditionConfig,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckLightstepAlertExists(resourceName, &condition),
					resource.TestCheckResourceAttr(resourceName, "name", "Too many requests"),
					resource.TestCheckResourceAttr(resourceName, "description", "A link to a playbook"),
					resource.TestCheckResourceAttr(resourceName, "query.0.query_string", uqlQuery+"\n"),
					resource.TestCheckResourceAttr(resourceName, "expression.0.is_no_data", "true"),
				),
			},
			{
				Config: updatedConditionConfig,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckLightstepAlertExists(resourceName, &condition),
					resource.TestCheckResourceAttr(resourceName, "name", "updated"),
					resource.TestCheckResourceAttr(resourceName, "description", "A link to a fresh playbook"),
					resource.TestCheckResourceAttr(resourceName, "expression.0.is_no_data", "false"),
				),
			},
		},
	})
}

func TestAccSpanLatencyAlert(t *testing.T) {
	var condition client.UnifiedCondition

	const uqlQuery = `spans latency | delta 1h, 1h | filter (service == "frontend") | group_by [], sum | point percentile(value, 50.0) | reduce 30s, min`

	conditionConfig := fmt.Sprintf(`
resource "lightstep_slack_destination" "slack" {
  project_name = "`+testProject+`"
  channel = "#emergency-room"
}

resource "lightstep_alert" "test" {
  project_name = "`+testProject+`"
  name = "Span latency alert"

  expression {
	  is_multi   = false
	  is_no_data = true
      operand  = "above"
	  thresholds {
		critical  = 10
		warning = 5
	  }
  }

  query {
    hidden              = false
    query_name          = "a"
    display = "line"
    query_string 	= <<EOT
%s
EOT
  }

  alerting_rule {
    id          = lightstep_slack_destination.slack.id
    update_interval = "1h"
  }
}
`, uqlQuery)

	const uqlQuery2 = `spans latency | delta 1h, 1h | filter (service == "frontend") | group_by [], sum | point percentile(value, 95.0) | reduce 30s, min`

	updatedConditionConfig := fmt.Sprintf(`
resource "lightstep_slack_destination" "slack" {
  project_name = "`+testProject+`"
  channel = "#emergency-room"
}

resource "lightstep_alert" "test" {
  project_name = "`+testProject+`"
  name = "Span latency alert - updated"

  expression {
	  is_multi   = false
	  is_no_data = true
      operand  = "above"
	  thresholds {
		critical  = 10
		warning = 5
	  }
  }

  query {
    hidden              = false
    query_name          = "a"
    display = "line"
    query_string 	= <<EOT
%s
EOT	
  }

  alerting_rule {
    id          = lightstep_slack_destination.slack.id
    update_interval = "1h"
  }
}
`, uqlQuery2)

	resourceName := "lightstep_alert.test"
	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccMetricConditionDestroy,
		Steps: []resource.TestStep{
			{
				Config: conditionConfig,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckLightstepAlertExists(resourceName, &condition),
					resource.TestCheckResourceAttr(resourceName, "name", "Span latency alert"),
					resource.TestCheckResourceAttr(resourceName, "query.0.query_string", uqlQuery+"\n"),
				),
			},
			{
				Config: updatedConditionConfig,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckLightstepAlertExists(resourceName, &condition),
					resource.TestCheckResourceAttr(resourceName, "name", "Span latency alert - updated"),
					resource.TestCheckResourceAttr(resourceName, "query.0.query_string", uqlQuery2+"\n"),
				),
			},
		},
	})
}

func TestAccAlertSpansQueryWithFormula(t *testing.T) {
	var condition client.UnifiedCondition

	uqlQuery := `spans latency | delta 1h | filter (service == "frontend") | group_by [], sum | point percentile(value, 50.0) | point (value + value)`

	conditionConfig := fmt.Sprintf(`
resource "lightstep_slack_destination" "slack" {
  project_name = "`+testProject+`"
  channel = "#emergency-room"
}

resource "lightstep_alert" "test" {
  project_name = "`+testProject+`"
  name = "Span rate alert"

  expression {
	  is_multi   = false
	  is_no_data = true
      operand  = "above"
	  thresholds {
		critical  = 10
		warning = 5
	  }
  }

  query {
    hidden              = false
    query_name          = "a"
    display = "line"
    query_string = <<EOT
%s
EOT
  }

  alerting_rule {
    id          = lightstep_slack_destination.slack.id
    update_interval = "1h"
  }
}
`, uqlQuery)

	uqlQuery2 := `spans latency | delta 1h | filter (service == "frontend") | group_by [], sum | point percentile(value, 50.0) | point (value + value + value)`

	updatedConditionConfig := fmt.Sprintf(`
resource "lightstep_slack_destination" "slack" {
  project_name = "`+testProject+`"
  channel = "#emergency-room"
}

resource "lightstep_alert" "test" {
  project_name = "`+testProject+`"
  name = "Span rate alert - updated"

  expression {
	  is_multi   = false
	  is_no_data = true
      operand  = "above"
	  thresholds {
		critical  = 10
		warning = 5
	  }
  }

  query {
    hidden     = false
    query_name = "a"
    display    = "line"
    query_string = <<EOT
%s
EOT
  }

  alerting_rule {
    id          = lightstep_slack_destination.slack.id
    update_interval = "1h"
  }
}
`, uqlQuery2)

	resourceName := "lightstep_alert.test"
	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccMetricConditionDestroy,
		Steps: []resource.TestStep{
			{
				Config: conditionConfig,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckLightstepAlertExists(resourceName, &condition),
					resource.TestCheckResourceAttr(resourceName, "name", "Span rate alert"),
					resource.TestCheckResourceAttr(resourceName, "query.0.query_string", uqlQuery+"\n"),
				),
			},
			{
				Config: updatedConditionConfig,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckLightstepAlertExists(resourceName, &condition),
					resource.TestCheckResourceAttr(resourceName, "query.0.query_string", uqlQuery2+"\n"),
				),
			},
		},
	})
}

func TestAccAlertWithFormula(t *testing.T) {
	var condition client.UnifiedCondition

	// The query string should come back exactly without losing the comment or user formatting
	const uqlQuery = `	
# Test comment to ensure this is retained too
with 
	a = metric requests | rate 1h | filter "service_name" == "frontend" | group_by ["method"], mean;
	b = metric requests | rate 1h | filter "service_name" == "frontend" && "error" == "true" | group_by ["method"], mean;
join (a/b)*100, a=0, b=0`
	conditionConfig := fmt.Sprintf(`
resource "lightstep_slack_destination" "slack" {
  project_name = "`+testProject+`"
  channel = "#emergency-room"
}

resource "lightstep_alert" "test" {
  project_name = "`+testProject+`"
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

  query {
	query_name                          = "a"
	hidden                              = false
    display                             = "line"
	query_string                        = <<EOT
%s
EOT
  }

  alerting_rule {
    id          = lightstep_slack_destination.slack.id
    update_interval = "1h"
  }
}
`, uqlQuery)

	resourceName := "lightstep_alert.test"
	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccMetricConditionDestroy,
		Steps: []resource.TestStep{
			{
				Config: conditionConfig,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckLightstepAlertExists(resourceName, &condition),
					resource.TestCheckResourceAttr(resourceName, "name", "Too many requests"),
					resource.TestCheckResourceAttr(resourceName, "query.0.query_string", uqlQuery+"\n"),
				),
			},
		},
	})
}

func TestAccCompositeAlert(t *testing.T) {
	var compositeCondition client.UnifiedCondition

	missingBothSingleAndCompositeFieldsCondition := `
resource "lightstep_alert" "errors" {
 project_name = "` + testProject + `"
 name = "no expression or query or composite alert"
 description = "elucidation..."
}
`

	includesBothSingleAndCompositeFieldsCondition := `
resource "lightstep_alert" "errors" {
 project_name = "` + testProject + `"
 name = "no expression or query or composite alert"
 description = "elucidation..."

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
	query_string        = "metric requests | rate 1h, 30s | filter \"project_name\" == \"catlab\" && \"service\" != \"android\" | group_by[\"method\"], mean | reduce 30s, min"
  }

 composite_alert {
     alert {
       name = "A"
       title = "Too many requests"
       expression {
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
	      query_string        = "metric requests | rate 1h, 30s | filter \"project_name\" == \"catlab\" && \"service\" != \"android\" | group_by[\"method\"], mean | reduce 30s, min"
       }
     }
     
     alert {
       name = "B"
       title = "Too many customers"
       expression {
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
	      query_string        = "metric customers | rate 1h, 30s | filter \"project_name\" == \"catlab\" && \"service\" != \"android\" | group_by[\"method\"], mean | reduce 30s, min"
       }
     }
 }

}
`

	compositeConditionConfig := `
	resource "lightstep_slack_destination" "slack" {
	project_name = "` + testProject + `"
	channel = "#emergency-room"
	}
	
	resource "lightstep_alert" "test" {
	project_name = "` + testProject + `"
	name = "Too many requests & customers"
	description = "A link to a playbook"
	
	composite_alert {
	    alert {
	      name = "A"
	      title = "Too many requests"
	      expression {
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
		      query_string        = "metric requests | rate 1h, 30s | filter \"project_name\" == \"catlab\" && \"service\" != \"android\" | group_by[\"method\"], mean | reduce 30s, min"
	      }
	    }
	
	    alert {
	      name = "B"
	      title = "Too many customers"
	      expression {
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
		      query_string        = "metric customers | rate 1h, 30s | filter \"project_name\" == \"catlab\" && \"service\" != \"android\" | group_by[\"method\"], mean | reduce 30s, min"
	      }
	    }
	}
	
	alerting_rule {
	  id          = lightstep_slack_destination.slack.id
	  update_interval = "1h"
	}
	}
	`

	updatedCompositeConditionConfig := `
resource "lightstep_slack_destination" "slack" {
 project_name = "` + testProject + `"
 channel = "#emergency-room"
}

resource "lightstep_alert" "test" {
 project_name = "` + testProject + `"
 name = "updated too many requests & customers"
 description = "A link to a playbook"

 composite_alert {
   alert {
       name = "A"
       title = "updated too many requests"
       expression {
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
	      query_string        = "metric requests | rate 1h, 30s | filter \"project_name\" == \"catlab\" && \"service\" != \"iOS\" | group_by[\"method\"], mean | reduce 30s, min"
       }
   }

   alert {
       name = "B"
       title = "updated too many customers"
       expression {
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
	      query_string        = "metric customers | rate 1h, 30s | filter \"project_name\" == \"catlab\" && \"service\" != \"iOS\" | group_by[\"method\"], mean | reduce 30s, min"
       }
   }
 }

 alerting_rule {
   id          = lightstep_slack_destination.slack.id
   update_interval = "1h"
 }
}
`

	noDataCompositeConditionConfig := `
resource "lightstep_slack_destination" "slack" {
 project_name = "` + testProject + `"
 channel = "#emergency-room"
}

resource "lightstep_alert" "test" {
 project_name = "` + testProject + `"
 name = "sub-alert A has no thresholds"
 description = "A link to a playbook"

 composite_alert {
   alert {
       name = "A"
       title = "no thresholds"
       expression {
	     is_no_data = true
         operand  = "above"
       }
       query {
         query_name          = "a"
         hidden              = false
         query_string        = "metric requests | rate 1h, 30s | filter \"project_name\" == \"catlab\" && \"service\" != \"iOS\" | group_by[\"method\"], mean | reduce 30s, min"
       }
   }

   alert {
       name = "B"
       title = "empty thresholds"
       expression {
	     is_no_data = true
         thresholds {}
       }
       query {
         query_name          = "a"
         hidden              = false
	     query_string        = "metric customers | rate 1h, 30s | filter \"project_name\" == \"catlab\" && \"service\" != \"iOS\" | group_by[\"method\"], mean | reduce 30s, min"
       }
   }
   alert {
       name = "C"
       title = "normal thresholds"
       expression {
	     is_no_data = false
         operand  = "above"
         thresholds {
	    	critical  = 20
	    	warning = 1
	      }
       }
       query {
         query_name          = "a"
         hidden              = false
	     query_string        = "metric customers | rate 1h, 30s | filter \"project_name\" == \"catlab\" && \"service\" != \"iOS\" | group_by[\"method\"], mean | reduce 30s, min"
       }
   }
 }

 alerting_rule {
   id          = lightstep_slack_destination.slack.id
   update_interval = "1h"
 }
}
`

	resourceName := "lightstep_alert.test"
	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccMetricConditionDestroy,
		Steps: []resource.TestStep{
			{
				Config: missingBothSingleAndCompositeFieldsCondition,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckLightstepAlertExists(resourceName, &compositeCondition),
				),
				ExpectError: regexp.MustCompile("at least one query is required"),
			},
			{
				Config: includesBothSingleAndCompositeFieldsCondition,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckLightstepAlertExists(resourceName, &compositeCondition),
				),
				ExpectError: regexp.MustCompile("single alert queries and composite alert cannot both be set"),
			},
			{
				Config: compositeConditionConfig,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckLightstepAlertExists(resourceName, &compositeCondition),
					resource.TestCheckResourceAttr(resourceName, "name", "Too many requests & customers"),
					resource.TestCheckResourceAttr(resourceName, "description", "A link to a playbook"),
					resource.TestCheckResourceAttr(resourceName, "composite_alert.0.alert.0.query.0.query_string", "metric requests | rate 1h, 30s | filter \"project_name\" == \"catlab\" && \"service\" != \"android\" | group_by[\"method\"], mean | reduce 30s, min"),
					resource.TestCheckResourceAttr(resourceName, "composite_alert.0.alert.0.expression.0.operand", "above"),
					resource.TestCheckResourceAttr(resourceName, "composite_alert.0.alert.0.expression.0.is_no_data", "false"),
					resource.TestCheckResourceAttr(resourceName, "composite_alert.0.alert.1.query.0.query_string", "metric customers | rate 1h, 30s | filter \"project_name\" == \"catlab\" && \"service\" != \"android\" | group_by[\"method\"], mean | reduce 30s, min"),
					resource.TestCheckResourceAttr(resourceName, "composite_alert.0.alert.1.expression.0.is_no_data", "false"),
					resource.TestCheckResourceAttr(resourceName, "composite_alert.0.alert.1.expression.0.thresholds.0.critical", "10"),
				),
			},
			{
				Config: updatedCompositeConditionConfig,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckLightstepAlertExists(resourceName, &compositeCondition),
					resource.TestCheckResourceAttr(resourceName, "name", "updated too many requests & customers"),
					resource.TestCheckResourceAttr(resourceName, "description", "A link to a playbook"),
					resource.TestCheckResourceAttr(resourceName, "composite_alert.0.alert.0.query.0.query_string", "metric requests | rate 1h, 30s | filter \"project_name\" == \"catlab\" && \"service\" != \"iOS\" | group_by[\"method\"], mean | reduce 30s, min"),
					resource.TestCheckResourceAttr(resourceName, "composite_alert.0.alert.0.expression.0.is_no_data", "true"),
					resource.TestCheckResourceAttr(resourceName, "composite_alert.0.alert.1.query.0.query_string", "metric customers | rate 1h, 30s | filter \"project_name\" == \"catlab\" && \"service\" != \"iOS\" | group_by[\"method\"], mean | reduce 30s, min"),
					resource.TestCheckResourceAttr(resourceName, "composite_alert.0.alert.1.expression.0.is_no_data", "true"),
				),
			},
			{
				Config: noDataCompositeConditionConfig,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckLightstepAlertExists(resourceName, &compositeCondition),
					resource.TestCheckResourceAttr(resourceName, "name", "sub-alert A has no thresholds"),
					resource.TestCheckResourceAttr(resourceName, "composite_alert.0.alert.0.name", "C"),
					resource.TestCheckResourceAttr(resourceName, "composite_alert.0.alert.0.expression.0.is_no_data", "false"),
					resource.TestCheckResourceAttr(resourceName, "composite_alert.0.alert.0.expression.0.thresholds.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "composite_alert.0.alert.0.expression.0.thresholds.0.critical", "20"),
					resource.TestCheckResourceAttr(resourceName, "composite_alert.0.alert.1.name", "B"),
					resource.TestCheckResourceAttr(resourceName, "composite_alert.0.alert.1.expression.0.is_no_data", "true"),
					resource.TestCheckResourceAttr(resourceName, "composite_alert.0.alert.1.expression.0.thresholds.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "composite_alert.0.alert.2.name", "A"),
					resource.TestCheckResourceAttr(resourceName, "composite_alert.0.alert.2.expression.0.is_no_data", "true"),
					resource.TestCheckResourceAttr(resourceName, "composite_alert.0.alert.2.expression.0.thresholds.#", "0"),
				),
			},
		},
	})
}

func testAccCheckLightstepAlertExists(resourceName string, condition *client.UnifiedCondition) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		tfCondition, ok := s.RootModule().Resources[resourceName]
		if !ok {
			return fmt.Errorf("not found: %s", resourceName)
		}

		if tfCondition.Primary.ID == "" {
			return fmt.Errorf("id is not set")
		}

		providerClient := testAccProvider.Meta().(*client.Client)
		cond, err := providerClient.GetUnifiedCondition(context.Background(), testProject, tfCondition.Primary.ID)
		if err != nil {
			return err
		}

		*condition = *cond
		return nil
	}
}

func TestAccEmptyUpdateInterval(t *testing.T) {
	var compositeCondition client.UnifiedCondition

	emptyUpdateIntervalConfig := `
resource "lightstep_slack_destination" "slack" {
 project_name = "` + testProject + `"
 channel = "#emergency-room"
}

resource "lightstep_alert" "test" {
 project_name = "` + testProject + `"
 name = "Too many requests & customers"
 description = "A link to a playbook"

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
	query_string        = "metric requests | rate 1h, 30s | filter \"project_name\" == \"catlab\" && \"service\" != \"iOS\" | group_by[\"method\"], mean | reduce 30s, min"
 }

 alerting_rule {
   id          = lightstep_slack_destination.slack.id
 }
}
`
	resourceName := "lightstep_alert.test"
	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccMetricConditionDestroy,
		Steps: []resource.TestStep{
			{
				Config: emptyUpdateIntervalConfig,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckLightstepAlertExists(resourceName, &compositeCondition),
					resource.TestCheckResourceAttr(resourceName, "name", "Too many requests & customers"),
					resource.TestCheckResourceAttr(resourceName, "description", "A link to a playbook"),
					resource.TestCheckResourceAttr(resourceName, "alerting_rule.0.update_interval", ""),
				),
			},
		},
	})
}

func TestAccUpdateCompositeAlertDiff_RegressionTest(t *testing.T) {
	var compositeCondition client.UnifiedCondition

	initialConfig := `
resource "lightstep_alert" "composite_diff_test" {
  project_name = "` + testProject + `"
  name         = "Composite Alert Diff"

  composite_alert {
    alert {
      name = "A"
      expression {
        is_no_data = false
        operand    = "above"
        thresholds {
          warning = 95
        }
      }
      query {
        hidden       = false
        query_name   = "a"
        query_string = "metric abc|rate 30s|group_by [], sum|point value/2"
	  }
    }

    alert {
      name = "B"
      expression {
        is_no_data = false
        operand    = "below"
        thresholds {
          warning = 97.000001
        }
      }
      query {
        hidden       = false
        query_name   = "b"
        query_string = "metric abc|rate 30s|group_by [], sum|point value/2"
	  }
    }
  }
}
`

	changeSubAlertQuery := `
resource "lightstep_alert" "composite_diff_test" {
  project_name = "` + testProject + `"
  name         = "Composite Alert Diff (updated query)"

  composite_alert {
    alert {
      name = "A"
      expression {
        is_no_data = false
        operand    = "above"
        thresholds {
          warning = 95
        }
      }
      query {
        hidden       = false
        query_name   = "a"
        query_string = "metric abc|rate 30s|group_by [], sum|point value/2"
	  }
    }

    alert {
      name = "B"
      expression {
        is_no_data = false
        operand    = "below"
        thresholds {

          warning = 97.000001
        }
      }
      query {
        hidden       = false
        query_name   = "c"
        query_string = "metric abc|rate 30s|group_by [], sum"
      }
    }
  }
}
`
	changeSubAlertName := `
resource "lightstep_alert" "composite_diff_test" {
  project_name = "` + testProject + `"
  name         = "Composite Alert Diff (updated sub-alert name)"

  composite_alert {
    alert {
      name = "A"
      expression {
        is_no_data = false
        operand    = "above"
        thresholds {
          warning = 95
        }
      }
      query {
        hidden       = false
        query_name   = "a"
        query_string = "metric abc|rate 30s|group_by [], sum|point value/2"
      }
    }

    alert {
      name = "C"
      expression {
        is_no_data = false
        operand    = "below"
        thresholds {

          warning = 97.000001
        }
      }
      query {
        hidden       = false
        query_name   = "c"
        query_string = "metric abc|rate 30s|group_by [], sum"
	  }
    }
  }
}
`

	resourceName := "lightstep_alert.composite_diff_test"
	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccMetricConditionDestroy,
		Steps: []resource.TestStep{
			{
				Destroy: false,
				Config:  initialConfig,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckLightstepAlertExists(resourceName, &compositeCondition),
					resource.TestCheckResourceAttr(resourceName, "name", "Composite Alert Diff"),
					resource.TestCheckResourceAttr(resourceName, "composite_alert.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "composite_alert.0.alert.#", "2"),
					resource.TestCheckResourceAttr(resourceName, "composite_alert.0.alert.0.name", "A"),
					resource.TestCheckResourceAttr(resourceName, "composite_alert.0.alert.0.query.0.query_string",
						"metric abc|rate 30s|group_by [], sum|point value/2"),
					resource.TestCheckResourceAttr(resourceName, "composite_alert.0.alert.1.name", "B"),
					resource.TestCheckResourceAttr(resourceName, "composite_alert.0.alert.1.query.0.query_string",
						"metric abc|rate 30s|group_by [], sum|point value/2"),
				),
			},
			{
				Config:  changeSubAlertQuery,
				Destroy: false,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckLightstepAlertExists(resourceName, &compositeCondition),
					resource.TestCheckResourceAttr(resourceName, "name", "Composite Alert Diff (updated query)"),
					resource.TestCheckResourceAttr(resourceName, "composite_alert.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "composite_alert.0.alert.#", "2"),
					resource.TestCheckResourceAttr(resourceName, "composite_alert.0.alert.0.name", "A"),
					resource.TestCheckResourceAttr(resourceName, "composite_alert.0.alert.0.query.0.query_string",
						"metric abc|rate 30s|group_by [], sum|point value/2"),
					resource.TestCheckResourceAttr(resourceName, "composite_alert.0.alert.1.name", "B"),
					resource.TestCheckResourceAttr(resourceName, "composite_alert.0.alert.1.query.0.query_string",
						"metric abc|rate 30s|group_by [], sum"),
				),
			},
			{
				Config:  changeSubAlertName,
				Destroy: false,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckLightstepAlertExists(resourceName, &compositeCondition),
					resource.TestCheckResourceAttr(resourceName, "name", "Composite Alert Diff (updated sub-alert name)"),
					resource.TestCheckResourceAttr(resourceName, "composite_alert.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "composite_alert.0.alert.#", "2"),
					resource.TestCheckResourceAttr(resourceName, "composite_alert.0.alert.0.name", "A"),
					resource.TestCheckResourceAttr(resourceName, "composite_alert.0.alert.0.query.0.query_string",
						"metric abc|rate 30s|group_by [], sum|point value/2"),
					resource.TestCheckResourceAttr(resourceName, "composite_alert.0.alert.1.name", "C"),
					resource.TestCheckResourceAttr(resourceName, "composite_alert.0.alert.1.query.0.query_string",
						"metric abc|rate 30s|group_by [], sum"),
				),
			},
		},
	})
}

func TestAccAlertWithThresholdDurations(t *testing.T) {
	var condition client.UnifiedCondition

	conditionConfig := `
resource "lightstep_alert" "test" {
  project_name = "` + testProject + `"
  name = "Too many requests"

  expression {
	  is_multi   = true
	  is_no_data = true
      no_data_duration_ms = 60000
      operand  = "above"
	  thresholds {
		critical  = 10
        critical_duration_ms = 180000
		warning = 5
        warning_duration_ms = 120000
	  }
  }

  query {
	query_name                          = "a"
	hidden                              = false
    display                             = "line"
	query_string                        = "metric requests | rate 1h | filter service_name == frontend | group_by [method], mean"
  }
}
`

	resourceName := "lightstep_alert.test"
	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccMetricConditionDestroy,
		Steps: []resource.TestStep{
			{
				Config: conditionConfig,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckLightstepAlertExists(resourceName, &condition),
					resource.TestCheckResourceAttr(resourceName, "name", "Too many requests"),
					resource.TestCheckResourceAttr(resourceName, "expression.0.no_data_duration_ms", "60000"),
					resource.TestCheckResourceAttr(resourceName, "expression.0.thresholds.0.warning_duration_ms", "120000"),
					resource.TestCheckResourceAttr(resourceName, "expression.0.thresholds.0.critical_duration_ms", "180000"),
				),
			},
		},
	})
}

func TestAccAlertWithOnlyWarningThresholdDurations(t *testing.T) {
	var condition client.UnifiedCondition

	conditionConfig := `
resource "lightstep_alert" "test" {
  project_name = "` + testProject + `"
  name = "Too many requests"

  expression {
	  is_multi   = true
	  is_no_data = true
      no_data_duration_ms = 60000
      operand  = "above"
	  thresholds {
		warning = 5
        warning_duration_ms = 120000
	  }
  }

  query {
	query_name                          = "a"
	hidden                              = false
    display                             = "line"
	query_string                        = "metric requests | rate 1h | filter service_name == frontend | group_by [method], mean"
  }
}
`

	resourceName := "lightstep_alert.test"
	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccMetricConditionDestroy,
		Steps: []resource.TestStep{
			{
				Config: conditionConfig,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckLightstepAlertExists(resourceName, &condition),
					resource.TestCheckResourceAttr(resourceName, "name", "Too many requests"),
					resource.TestCheckResourceAttr(resourceName, "expression.0.no_data_duration_ms", "60000"),
					resource.TestCheckResourceAttr(resourceName, "expression.0.thresholds.0.warning_duration_ms", "120000"),
				),
			},
		},
	})
}

func TestAccAlertWithOnlyCriticalThresholdDurations(t *testing.T) {
	var condition client.UnifiedCondition

	conditionConfig := `
resource "lightstep_alert" "test" {
  project_name = "` + testProject + `"
  name = "Too many requests"

  expression {
	  is_multi   = true
	  is_no_data = true
      no_data_duration_ms = 60000
      operand  = "above"
	  thresholds {
		critical  = 10
        critical_duration_ms = 180000
	  }
  }

  query {
	query_name                          = "a"
	hidden                              = false
    display                             = "line"
	query_string                        = "metric requests | rate 1h | filter service_name == frontend | group_by [method], mean"
  }
}
`

	resourceName := "lightstep_alert.test"
	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccMetricConditionDestroy,
		Steps: []resource.TestStep{
			{
				Config: conditionConfig,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckLightstepAlertExists(resourceName, &condition),
					resource.TestCheckResourceAttr(resourceName, "name", "Too many requests"),
					resource.TestCheckResourceAttr(resourceName, "expression.0.no_data_duration_ms", "60000"),
					resource.TestCheckResourceAttr(resourceName, "expression.0.thresholds.0.critical_duration_ms", "180000"),
				),
			},
		},
	})
}

func TestAlertWithHiddenQueries(t *testing.T) {
	var condition client.UnifiedCondition

	conditionConfig := `
resource "lightstep_alert" "test" {
	name         = "hidden queries spurious diff"
	project_name = "` + testProject + `"
	
	expression {
		operand = "below"
		thresholds {
			critical = "1"
		}
	}
	
	query {
		display = "line"
		hidden = false
		// previously, this would result in a spurious diff
		hidden_queries = {
			b = true
		}
		query_name = "a"
		query_string = "metric cpu.utilization | delta 5m | group_by[], sum"
	}
}
`

	resourceName := "lightstep_alert.test"
	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccMetricConditionDestroy,
		Steps: []resource.TestStep{
			{
				Config: conditionConfig,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckLightstepAlertExists(resourceName, &condition),
					resource.TestCheckResourceAttr(resourceName, "name", "hidden queries spurious diff"),
					resource.TestCheckResourceAttr(resourceName, "query.0.hidden_queries.%", "1"),
					resource.TestCheckResourceAttr(resourceName, "query.0.hidden_queries.b", "true"),
				),
			},
		},
	})
}

func TestCompositeCrash(t *testing.T) {
	var condition client.UnifiedCondition

	conditionConfig := `
resource "lightstep_alert" "test" {
  name         = "test"
  project_name = "` + testProject + `"

  composite_alert {

    alert {
      name  = "A"
      title = "cpu"

      query {
        hidden       = false
        query_name   = "a"
        query_string = "metric a | delta 30s | group_by[], sum"
        display      = "line"
        hidden_queries = { "a" = "%s" }
      }

      expression {
        operand = "above"

        thresholds {
          critical = 20
        }
      }
    }
	alert {
      name  = "B"
      title = "b"

      query {
        hidden       = false
        query_name   = "a"
        query_string = "metric b | delta 30s | group_by[], sum"
        display      = "line" 
      }

      expression {
        operand = "above"

        thresholds {
          critical = 20
        }
      }
    }
  }
}
`

	resourceName := "lightstep_alert.test"
	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccMetricConditionDestroy,
		Steps: []resource.TestStep{
			{
				Config: fmt.Sprintf(conditionConfig, "false"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckLightstepAlertExists(resourceName, &condition),
				),
			},
			{
				Config: fmt.Sprintf(conditionConfig, "true"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckLightstepAlertExists(resourceName, &condition),
				),
			},
		},
	})
}
