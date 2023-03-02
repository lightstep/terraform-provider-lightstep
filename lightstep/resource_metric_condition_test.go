package lightstep

import (
	"context"
	"fmt"
	"reflect"
	"regexp"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"

	"github.com/lightstep/terraform-provider-lightstep/client"

	"github.com/stretchr/testify/require"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
)

var (
	k = "key"
	v = "value"

	includeFilter = client.LabelFilter{
		Key:     k,
		Value:   v,
		Operand: "eq",
	}

	excludeFilter = client.LabelFilter{
		Key:     k,
		Value:   v,
		Operand: "neq",
	}

	allFilter = client.LabelFilter{
		Key:     k,
		Value:   v,
		Operand: "contains",
	}
)

func TestAccMetricCondition(t *testing.T) {
	var condition client.UnifiedCondition

	badCondition := `
resource "lightstep_metric_condition" "errors" {
  project_name = "terraform-provider-tests"
  name = "Too many requests"
  expression {
	  evaluation_window   = "2m"
	  evaluation_criteria = "on_average"
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

resource "lightstep_metric_condition" "test" {
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

  metric_query {
    metric         = "requests"
    query_name          = "a"
    timeseries_operator = "rate"
    timeseries_operator_input_window_ms = 3600000
    hidden              = false
    display = "line"
    include_filters = [{
      key   = "project_name"
      value = "catlab"
    }]

    exclude_filters = [{
      key   = "service"
      value = "android"
    }]

    group_by  {
      aggregation_method = "avg"
      keys = ["method"]
    }

    final_window_operation {
      operator = "min"
      input_window_ms  = 30000
    }
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

resource "lightstep_metric_condition" "test" {
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

  metric_query {
    metric         = "requests"
    query_name          = "a"
    timeseries_operator = "rate"
    timeseries_operator_input_window_ms = 3600000
    hidden              = false
	display             = "line"

    include_filters = [{
      key   = "project_name"
      value = "catlab"
    }]

    exclude_filters = [{
      key   = "service"
      value = "android"
    }]

    group_by  {
      aggregation_method = "avg"
      keys = ["method"]
    }
    
    final_window_operation {
      operator = "min"
      input_window_ms  = 30000
    }
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

	resourceName := "lightstep_metric_condition.test"
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
				ExpectError: regexp.MustCompile("(Missing required argument|Insufficient metric_query blocks)"),
			},
			{
				Config: conditionConfig,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckMetricConditionExists(resourceName, &condition),
					resource.TestCheckResourceAttr(resourceName, "name", "Too many requests"),
					resource.TestCheckResourceAttr(resourceName, "description", "A link to a playbook"),
					resource.TestCheckResourceAttr(resourceName, "metric_query.0.tql", "metric requests | filter ((project_name == \"catlab\") && (service != \"android\")) | rate 1h, 1h | group_by [\"method\"], mean | reduce 30s, min"),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "alerting_rule.*", map[string]string{
						"include_filters.0.key":   "project_name",
						"include_filters.0.value": "catlab",
						"filters.0.key":           "service_name",
						"filters.0.operand":       "contains",
						"filters.0.value":         "frontend",
					}),
					resource.TestCheckResourceAttr(resourceName, "expression.0.is_no_data", "true"),
				),
				ExpectNonEmptyPlan: true,
			},
			{
				Config: updatedConditionConfig,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckMetricConditionExists(resourceName, &condition),
					resource.TestCheckResourceAttr(resourceName, "name", "updated"),
					resource.TestCheckResourceAttr(resourceName, "description", "A link to a fresh playbook"),
					resource.TestCheckResourceAttr(resourceName, "expression.0.is_no_data", "false"),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func TestAccSpanLatencyCondition(t *testing.T) {
	var condition client.UnifiedCondition

	conditionConfig := `
resource "lightstep_slack_destination" "slack" {
  project_name = "terraform-provider-tests"
  channel = "#emergency-room"
}

resource "lightstep_metric_condition" "test" {
  project_name = "terraform-provider-tests"
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

  metric_query {
    hidden              = false
    query_name          = "a"
    display = "line"
    spans {
      query = "service IN (\"frontend\")"
      operator = "latency"
      operator_input_window_ms = 3600000
      latency_percentiles = [50]
    }

    final_window_operation {
      operator = "min"
      input_window_ms  = 30000
    }
  }

  alerting_rule {
    id          = lightstep_slack_destination.slack.id
    update_interval = "1h"
  }
}
`

	updatedConditionConfig := `
resource "lightstep_slack_destination" "slack" {
  project_name = "terraform-provider-tests"
  channel = "#emergency-room"
}

resource "lightstep_metric_condition" "test" {
  project_name = "terraform-provider-tests"
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

  metric_query {
    hidden              = false
    query_name          = "a"
    display = "line"
    spans {
      query = "service IN (\"frontend\")"
      operator = "latency"
      operator_input_window_ms = 3600000
      latency_percentiles = [95]
    }

    final_window_operation {
      operator = "min"
      input_window_ms  = 30000
    }
  }

  alerting_rule {
    id          = lightstep_slack_destination.slack.id
    update_interval = "1h"
  }
}
`

	resourceName := "lightstep_metric_condition.test"
	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccMetricConditionDestroy,
		Steps: []resource.TestStep{
			{
				Config: conditionConfig,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckMetricConditionExists(resourceName, &condition),
					resource.TestCheckResourceAttr(resourceName, "name", "Span latency alert"),
					resource.TestCheckResourceAttr(resourceName, "metric_query.0.tql", "spans latency | delta 1h, 1h | filter (service == \"frontend\") | group_by [], sum | point percentile(value, 50.0) | reduce 30s, min"),
				),
				ExpectNonEmptyPlan: true,
			},
			{
				Config: updatedConditionConfig,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckMetricConditionExists(resourceName, &condition),
					resource.TestCheckResourceAttr(resourceName, "name", "Span latency alert - updated"),
					resource.TestCheckResourceAttr(resourceName, "metric_query.0.tql", "spans latency | delta 1h, 1h | filter (service == \"frontend\") | group_by [], sum | point percentile(value, 50.0) | reduce 30s, min"),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func TestAccSpanRateCondition(t *testing.T) {
	var condition client.UnifiedCondition

	conditionConfig := `
resource "lightstep_slack_destination" "slack" {
  project_name = "terraform-provider-tests"
  channel = "#emergency-room"
}

resource "lightstep_metric_condition" "test" {
  project_name = "terraform-provider-tests"
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

  metric_query {
    hidden              = false
    query_name          = "a"
    display = "line"
    spans {
      query = "service IN (\"frontend\")"
      operator = "rate"
      operator_input_window_ms = 3600000
    }

    final_window_operation {
      operator = "min"
      input_window_ms  = 30000
    }
  }

  alerting_rule {
    id          = lightstep_slack_destination.slack.id
    update_interval = "1h"
  }
}
`

	updatedConditionConfig := `
resource "lightstep_slack_destination" "slack" {
  project_name = "terraform-provider-tests"
  channel = "#emergency-room"
}

resource "lightstep_metric_condition" "test" {
  project_name = "terraform-provider-tests"
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

  metric_query {
    hidden              = false
    query_name          = "a"
    display = "line"
    spans {
      query = "service IN (\"frontend\")"
      operator = "rate"
      operator_input_window_ms = 3600000
    }

    final_window_operation {
      operator = "min"
      input_window_ms  = 30000
    }
  }

  alerting_rule {
    id          = lightstep_slack_destination.slack.id
    update_interval = "1h"
  }
}
`

	resourceName := "lightstep_metric_condition.test"
	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccMetricConditionDestroy,
		Steps: []resource.TestStep{
			{
				Config: conditionConfig,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckMetricConditionExists(resourceName, &condition),
					resource.TestCheckResourceAttr(resourceName, "name", "Span rate alert"),
					resource.TestCheckResourceAttr(resourceName, "metric_query.0.tql", "spans count | rate 1h, 1h | filter (service == \"frontend\") | group_by [], sum | reduce 30s, min"),
				),
				ExpectNonEmptyPlan: true,
			},
			{
				Config: updatedConditionConfig,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckMetricConditionExists(resourceName, &condition),
					resource.TestCheckResourceAttr(resourceName, "name", "Span rate alert - updated"),
					resource.TestCheckResourceAttr(resourceName, "metric_query.0.tql", "spans count | rate 1h, 1h | filter (service == \"frontend\") | group_by [], sum | reduce 30s, min"),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func TestAccSpanErrorRatioCondition(t *testing.T) {
	var condition client.UnifiedCondition

	conditionConfig := `
resource "lightstep_slack_destination" "slack" {
  project_name = "terraform-provider-tests"
  channel = "#emergency-room"
}

resource "lightstep_metric_condition" "test" {
  project_name = "terraform-provider-tests"
  name = "Span error ratio alert"

  expression {
	  is_multi   = false
	  is_no_data = true
      operand  = "above"
	  thresholds {
		critical  = 0.9
		warning = 0.5
	  }
  }

  metric_query {
    hidden              = false
    query_name          = "a"
    display = "line"
    spans {
      query = "service IN (\"frontend\")"
      operator = "error_ratio"
      operator_input_window_ms = 3600000
    }

    final_window_operation {
      operator = "min"
      input_window_ms  = 30000
    }
  }

  alerting_rule {
    id          = lightstep_slack_destination.slack.id
    update_interval = "1h"
  }
}
`

	updatedConditionConfig := `
resource "lightstep_slack_destination" "slack" {
  project_name = "terraform-provider-tests"
  channel = "#emergency-room"
}

resource "lightstep_metric_condition" "test" {
  project_name = "terraform-provider-tests"
  name = "Span error ratio alert - updated"

  expression {
	  is_multi   = false
	  is_no_data = true
      operand  = "above"
	  thresholds {
		critical  = 0.9
		warning = 0.5
	  }
  }

  metric_query {
    hidden              = false
    query_name          = "a"
    display = "line"
    spans {
      query = "service IN (\"frontend\")"
      operator = "error_ratio"
      operator_input_window_ms = 3600000
    }

    final_window_operation {
      operator = "min"
      input_window_ms  = 30000
    }
  }

  alerting_rule {
    id          = lightstep_slack_destination.slack.id
    update_interval = "1h"
  }
}
`

	resourceName := "lightstep_metric_condition.test"
	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccMetricConditionDestroy,
		Steps: []resource.TestStep{
			{
				Config: conditionConfig,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckMetricConditionExists(resourceName, &condition),
					resource.TestCheckResourceAttr(resourceName, "name", "Span error ratio alert"),
					resource.TestCheckResourceAttr(resourceName, "metric_query.0.tql", "with\n  errors = spans count | delta 1h, 1h | filter ((service == \"frontend\") && (error == true)) | group_by [], sum;\n  total = spans count | delta 1h, 1h | filter (service == \"frontend\") | group_by [], sum;\njoin (errors / total), errors=0, total=0 | reduce 30s, min"),
				),
				// This is expected to conver the above query definition to a query string
				ExpectNonEmptyPlan: true,
			},
			{
				Config: updatedConditionConfig,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckMetricConditionExists(resourceName, &condition),
					resource.TestCheckResourceAttr(resourceName, "name", "Span error ratio alert - updated"),
					resource.TestCheckResourceAttr(resourceName, "metric_query.0.tql", "with\n  errors = spans count | delta 1h, 1h | filter ((service == \"frontend\") && (error == true)) | group_by [], sum;\n  total = spans count | delta 1h, 1h | filter (service == \"frontend\") | group_by [], sum;\njoin (errors / total), errors=0, total=0 | reduce 30s, min"),
				),
				// This is expected to conver the above query definition to a query string
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func TestAccSpanRateConditionWithFormula(t *testing.T) {
	var condition client.UnifiedCondition

	conditionConfig := `
resource "lightstep_slack_destination" "slack" {
  project_name = "terraform-provider-tests"
  channel = "#emergency-room"
}

resource "lightstep_metric_condition" "test" {
  project_name = "terraform-provider-tests"
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

  metric_query {
    hidden              = false
    query_name          = "a"
    display = "line"
    spans {
      query = "service IN (\"frontend\")"
      operator = "rate"
      operator_input_window_ms = 3600000
    }
  }

  metric_query {
    hidden              = false
    query_name          = "a+a"
    display = "line"

    final_window_operation {
      operator = "min"
      input_window_ms  = 30000
    }
  }

  alerting_rule {
    id          = lightstep_slack_destination.slack.id
    update_interval = "1h"
  }
}
`

	updatedConditionConfig := `
resource "lightstep_slack_destination" "slack" {
  project_name = "terraform-provider-tests"
  channel = "#emergency-room"
}

resource "lightstep_metric_condition" "test" {
  project_name = "terraform-provider-tests"
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

  metric_query {
    hidden     = false
    query_name = "a"
    display    = "line"
    spans {
      query = "service IN (\"frontend\")"
      operator = "rate"
      operator_input_window_ms = 3600000
    }
  }

  metric_query {
    hidden     = false
    query_name = "a+a+a"
    display    = "line"

    final_window_operation {
      operator = "min"
      input_window_ms  = 30000
    }
  }

  alerting_rule {
    id          = lightstep_slack_destination.slack.id
    update_interval = "1h"
  }
}
`

	resourceName := "lightstep_metric_condition.test"
	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccMetricConditionDestroy,
		Steps: []resource.TestStep{
			{
				Config: conditionConfig,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckMetricConditionExists(resourceName, &condition),
					resource.TestCheckResourceAttr(resourceName, "name", "Span rate alert"),
					resource.TestCheckResourceAttr(resourceName, "metric_query.0.tql", "spans count | rate 1h, 1h | filter (service == \"frontend\") | group_by [], sum | point (value + value) | reduce 30s, min"),
				),
				ExpectNonEmptyPlan: true,
			},
			{
				Config: updatedConditionConfig,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckMetricConditionExists(resourceName, &condition),
					resource.TestCheckResourceAttr(resourceName, "metric_query.0.tql", "spans count | rate 1h, 1h | filter (service == \"frontend\") | group_by [], sum | point (value + value) | reduce 30s, min"),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func TestAccMetricConditionWithFormula(t *testing.T) {
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
  project_name = "terraform-provider-tests"
  channel = "#emergency-room"
}

resource "lightstep_alert" "test" {
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

    include_filters = [{
      key   = "project_name"
      value = "catlab"
    }]

	filters = [{
		  key   = "service_name"
		  value = "frontend"
		  operand = "contains"
	}]
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
					testAccCheckMetricConditionExists(resourceName, &condition),
					resource.TestCheckResourceAttr(resourceName, "name", "Too many requests"),
					resource.TestCheckResourceAttr(resourceName, "query.0.query_string", uqlQuery+"\n"),
				),
			},
		},
	})
}

func testAccCheckMetricConditionExists(resourceName string, condition *client.UnifiedCondition) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		tfCondition, ok := s.RootModule().Resources[resourceName]
		if !ok {
			return fmt.Errorf("not found: %s", resourceName)
		}

		if tfCondition.Primary.ID == "" {
			return fmt.Errorf("id is not set")
		}

		providerClient := testAccProvider.Meta().(*client.Client)
		cond, err := providerClient.GetUnifiedCondition(context.Background(), test_project, tfCondition.Primary.ID)
		if err != nil {
			return err
		}

		condition = cond
		return nil
	}
}

func testAccMetricConditionDestroy(s *terraform.State) error {
	conn := testAccProvider.Meta().(*client.Client)
	for _, res := range s.RootModule().Resources {
		if res.Type != "metric_alert" {
			continue
		}

		s, err := conn.GetUnifiedCondition(context.Background(), test_project, res.Primary.ID)
		if err == nil {
			if s.ID == res.Primary.ID {
				return fmt.Errorf("metric condition with ID (%v) still exists.", res.Primary.ID)
			}
		}
	}
	return nil
}

func TestBuildLabelFilters(t *testing.T) {
	type filtersCase struct {
		includes []interface{}
		excludes []interface{}
		all      []interface{}
		expected []client.LabelFilter
	}

	cases := []filtersCase{
		// valid empty includes, valid excludes
		{
			includes: []interface{}{},
			excludes: []interface{}{
				map[string]interface{}{
					"key":   k,
					"value": v,
				},
			},
			expected: []client.LabelFilter{
				excludeFilter,
			},
		},
		// valid includes, valid empty excludes
		{
			includes: []interface{}{
				map[string]interface{}{
					"key":   k,
					"value": v,
				},
			},
			excludes: []interface{}{},
			expected: []client.LabelFilter{
				includeFilter,
			},
		},
		// valid includes valid excludes
		{
			includes: []interface{}{
				map[string]interface{}{
					"key":   k,
					"value": v,
				},
			},
			excludes: []interface{}{
				map[string]interface{}{
					"key":   k,
					"value": v,
				},
			},
			expected: []client.LabelFilter{
				includeFilter,
				excludeFilter,
			},
		},
		{
			all: []interface{}{
				map[string]interface{}{
					"key":     k,
					"operand": "contains",
					"value":   v,
				},
			},
			includes: []interface{}{},
			excludes: []interface{}{},
			expected: []client.LabelFilter{
				allFilter,
			},
		},
	}

	for _, c := range cases {
		result := buildLabelFilters(c.includes, c.excludes, c.all)
		require.Equal(t, c.expected, result)
	}
}

func TestBuildAlertingRules(t *testing.T) {
	type alertingRuleCase struct {
		rules    []interface{}
		expected []client.AlertingRule
	}

	id := "abc123"
	renotify := "1h"
	renotifyMillis := 3600000

	cases := []alertingRuleCase{
		// without includes or excludes
		{
			rules: []interface{}{
				map[string]interface{}{
					"id":              id,
					"update_interval": renotify,
				},
			},
			expected: []client.AlertingRule{
				{
					MessageDestinationID: id,
					UpdateInterval:       renotifyMillis,
				},
			},
		},
		// with includes
		{
			rules: []interface{}{
				map[string]interface{}{
					"id":              id,
					"update_interval": renotify,
					"include_filters": []interface{}{
						map[string]interface{}{
							"key":   k,
							"value": v,
						},
					},
				},
			},
			expected: []client.AlertingRule{
				{
					MessageDestinationID: id,
					UpdateInterval:       renotifyMillis,
					MatchOn:              client.MatchOn{GroupBy: []client.LabelFilter{includeFilter}},
				},
			},
		},
		// with excludes
		{
			rules: []interface{}{
				map[string]interface{}{
					"id":              id,
					"update_interval": renotify,
					"exclude_filters": []interface{}{
						map[string]interface{}{
							"key":   k,
							"value": v,
						},
					},
				},
			},
			expected: []client.AlertingRule{
				{
					MessageDestinationID: id,
					UpdateInterval:       renotifyMillis,
					MatchOn:              client.MatchOn{GroupBy: []client.LabelFilter{excludeFilter}},
				},
			},
		},
		// with both includes excludes
		{
			rules: []interface{}{
				map[string]interface{}{
					"id":              id,
					"update_interval": renotify,
					"include_filters": []interface{}{
						map[string]interface{}{
							"key":   k,
							"value": v,
						},
					},
					"exclude_filters": []interface{}{
						map[string]interface{}{
							"key":   k,
							"value": v,
						},
					},
				},
			},
			expected: []client.AlertingRule{
				{
					MessageDestinationID: id,
					UpdateInterval:       renotifyMillis,
					MatchOn:              client.MatchOn{GroupBy: []client.LabelFilter{includeFilter, excludeFilter}},
				},
			},
		},
	}

	for _, c := range cases {
		alertingRuleSet := schema.NewSet(
			schema.HashResource(&schema.Resource{
				Schema: getAlertingRuleSchema(),
			}),
			c.rules,
		)
		result, err := buildAlertingRules(alertingRuleSet)
		require.NoError(t, err)
		require.Equal(t, c.expected, result)
		require.Equal(t, c.expected, result)
	}
}

func TestValidateFilters(t *testing.T) {
	type filterCase struct {
		filters    []interface{}
		expectErr  bool
		hasOperand bool
	}

	cases := []filterCase{
		// has valid key and valid value
		{
			filters: []interface{}{
				map[string]interface{}{
					"key":   "key1",
					"value": "value1",
				},
			},
			expectErr:  false,
			hasOperand: false,
		},
		// has valid key and valid value and valid operand
		{
			filters: []interface{}{
				map[string]interface{}{
					"key":     "key1",
					"value":   "value1",
					"operand": "contains",
				},
			},
			expectErr:  false,
			hasOperand: true,
		},
		// missing value
		{
			filters: []interface{}{
				map[string]interface{}{
					"key": "key1",
				},
			},
			expectErr:  true,
			hasOperand: false,
		},
		// missing key
		{
			filters: []interface{}{
				map[string]interface{}{
					"value": "value1",
				},
			},
			expectErr:  true,
			hasOperand: false,
		},
		// key is not a string
		{
			filters: []interface{}{
				map[string]interface{}{
					"key":   1,
					"value": "some-value",
				},
			},
			expectErr:  true,
			hasOperand: false,
		},
		// value is not a string
		{
			filters: []interface{}{
				map[string]interface{}{
					"key":   "some-key",
					"value": 1,
				},
			},
			expectErr:  true,
			hasOperand: false,
		},
		// operand value is not a string
		{
			filters: []interface{}{
				map[string]interface{}{
					"key":     "some-key",
					"value":   "some-val",
					"operand": 1,
				},
			},
			expectErr:  true,
			hasOperand: true,
		},
		// operand value is eq
		{
			filters: []interface{}{
				map[string]interface{}{
					"key":     "some-key",
					"value":   "some-val",
					"operand": "eq",
				},
			},
			expectErr:  true,
			hasOperand: true,
		},
		// operand value is eq
		{
			filters: []interface{}{
				map[string]interface{}{
					"key":     "some-key",
					"value":   "some-val",
					"operand": "neq",
				},
			},
			expectErr:  true,
			hasOperand: true,
		},
	}

	for _, c := range cases {
		err := validateFilters(c.filters, c.hasOperand)
		if c.expectErr {
			require.Error(t, err)
		} else {
			require.NoError(t, err)
		}
	}
}

func TestValidateGroupBy(t *testing.T) {
	const (
		single    = "single"
		composite = "composite"
	)
	type groupByCase struct {
		groupBy   interface{}
		queryType string
		expectErr bool
	}

	cases := []groupByCase{
		// single, nil groupBy
		{
			groupBy:   nil,
			queryType: single,
			expectErr: true,
		},
		// single, empty groupBy
		{
			groupBy:   []interface{}{},
			queryType: single,
			expectErr: true,
		},
		// single, missing groupBy aggregation_method
		{
			groupBy: []interface{}{
				map[string]interface{}{
					"keys": []string{"key1"},
				},
			},
			queryType: single,
			expectErr: true,
		},
		// single, no keys (valid)
		{
			groupBy: []interface{}{
				map[string]interface{}{
					"aggregation_method": "avg",
				},
			},
			queryType: single,
			expectErr: false,
		},
		// composite, nil groupBy
		{
			groupBy:   nil,
			queryType: composite,
			expectErr: false,
		},
		// composite, empty groupBy
		{
			groupBy:   []interface{}{},
			queryType: composite,
			expectErr: false,
		},
		// composite, with groupBy no keys
		{
			groupBy: []interface{}{
				map[string]interface{}{
					"aggregation_method": "avg",
				},
			},
			queryType: composite,
			expectErr: true,
		},
		// composite, with groupBy
		{
			groupBy: []interface{}{
				map[string]interface{}{
					"aggregation_method": "avg",
					"keys":               []string{"key1"},
				},
			},
			queryType: composite,
			expectErr: true,
		},
	}

	for _, c := range cases {
		err := validateGroupBy(c.groupBy, c.queryType)
		if c.expectErr {
			require.Error(t, err)
		} else {
			require.NoError(t, err)
		}
	}
}

func Test_buildLatencyPercentiles(t *testing.T) {
	type args struct {
		lats    []interface{}
		display string
	}
	tests := []struct {
		name string
		args args
		want []float64
	}{
		{
			name: "Full List Provided; Expect Full List",
			args: args{
				lats:    []interface{}{float64(50), float64(95), float64(99), 99.9},
				display: "line",
			},
			want: []float64{50, 95, 99, 99.9},
		},
		{
			name: "Partial List Provided; Expect Partial List",
			args: args{
				lats:    []interface{}{float64(50), float64(95)},
				display: "line",
			},
			want: []float64{50, 95},
		},
		{
			name: "No List Provided; Expect Full List",
			args: args{
				lats:    []interface{}{},
				display: "line",
			},
			want: []float64{50, 95, 99, 99.9},
		},
		{
			name: "Heatmap: Full List Provided; Expect No List",
			args: args{
				lats:    []interface{}{float64(50), float64(95), float64(99), 99.9},
				display: "heatmap",
			},
			want: []float64{},
		},
		{
			name: "Heatmap: Partial List Provided; Expect No List",
			args: args{
				lats:    []interface{}{float64(50), float64(95)},
				display: "heatmap",
			},
			want: []float64{},
		},
		{
			name: "Heatmap: No List Provided; Expect No List",
			args: args{
				lats:    []interface{}{},
				display: "heatmap",
			},
			want: []float64{},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := buildLatencyPercentiles(tt.args.lats, tt.args.display); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("buildLatencyPercentiles() = %v, want %v", got, tt.want)
			}
		})
	}
}
