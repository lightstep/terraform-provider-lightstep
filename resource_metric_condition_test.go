package main

import (
	"fmt"
	"regexp"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/hashicorp/terraform-plugin-sdk/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/terraform"
	"github.com/lightstep/terraform-provider-lightstep/lightstep"
)

var (
	k = "key"
	v = "value"

	includeFilter = lightstep.LabelFilter{
		Key:     k,
		Value:   v,
		Operand: "eq",
	}

	excludeFilter = lightstep.LabelFilter{
		Key:     k,
		Value:   v,
		Operand: "neq",
	}
)

func TestAccMetricCondition(t *testing.T) {
	var condition lightstep.MetricCondition

	badCondition := `
resource "lightstep_metric_condition" "errors" {
  project_name = "terraform-provider-tests"
  condition_name = "Too many requests"

  evaluation_window   = "2m" 
  evaluation_criteria = "on_average"

  display = "line"

  is_multi   = true
  is_no_data = true

  thresholds = {
    operand  = "above"
    critical  = 10
    warning = 5
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
  condition_name = "Too many requests"

  evaluation_window   = "2m" 
  evaluation_criteria = "on_average"

  display = "line"

  is_multi   = true
  is_no_data = true

  thresholds = {
    operand  = "above"
    critical  = 10
    warning = 5
  }

  query {
    metric_name         = "requests"
    query_name          = "a"
    type                = "single"
    timeseries_operator = "rate"
    hidden              = false

    include_filters = [{
      key   = "project_name"
      value = "catlab"
    }]

    exclude_filters = [{
      key   = "service"
      value = "android"
    }]

    group_by  {
      aggregation = "avg"
      keys = ["method"]
    }
  }

  alerting_rule {
    id          = lightstep_slack_destination.slack.id
    renotify = "1h"

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
				ExpectError: regexp.MustCompile("required field is not set"),
			},
			{
				Config: conditionConfig,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckMetricConditionExists(resourceName, &condition),
					resource.TestCheckResourceAttr(resourceName, "condition_name", "Too many requests"),
					resource.TestCheckResourceAttr(resourceName, "evaluation_window", "2m"),
					resource.TestCheckResourceAttr(resourceName, "evaluation_criteria", "on_average"),
					resource.TestCheckResourceAttr(resourceName, "display", "line"),
					resource.TestCheckResourceAttr(resourceName, "is_multi", "true"),
					resource.TestCheckResourceAttr(resourceName, "is_no_data", "true"),
				),
			},
		},
	})
}

func testAccCheckMetricConditionExists(resourceName string, condition *lightstep.MetricCondition) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		tfCondition, ok := s.RootModule().Resources[resourceName]
		if !ok {
			return fmt.Errorf("Not found: %s", resourceName)
		}

		if tfCondition.Primary.ID == "" {
			return fmt.Errorf("ID is not set")
		}

		client := testAccProvider.Meta().(*lightstep.Client)
		cond, err := client.GetMetricCondition(test_project, tfCondition.Primary.ID)
		if err != nil {
			return err
		}

		*condition = cond
		return nil
	}
}

// confirms conditions created for test have been destroyed
func testAccMetricConditionDestroy(s *terraform.State) error {
	conn := testAccProvider.Meta().(*lightstep.Client)
	for _, resource := range s.RootModule().Resources {
		if resource.Type != "metric_alert" {
			continue
		}

		s, err := conn.GetMetricCondition(test_project, resource.Primary.ID)
		if err == nil {
			if s.ID == resource.Primary.ID {
				return fmt.Errorf("Metric condition with ID (%v) still exists.", resource.Primary.ID)
			}
		}
	}
	return nil
}

func TestBuildThresholds(t *testing.T) {
	type thresholdCase struct {
		thresholds map[string]interface{}
		expected   lightstep.Thresholds
		shouldErr  bool
	}

	cases := []thresholdCase{
		// valid critical
		{
			thresholds: map[string]interface{}{
				"critical": "5",
			},
			expected: lightstep.Thresholds{
				Critical: 5,
			},
			shouldErr: false,
		},
		// valid critical and warning
		{
			thresholds: map[string]interface{}{
				"critical": "5",
				"warning":  "10",
			},
			expected: lightstep.Thresholds{
				Critical: 5,
				Warning:  10,
			},
			shouldErr: false,
		},
		// valid critical, invalid warning
		{
			thresholds: map[string]interface{}{
				"critical": "5",
				"warning":  10,
			},
			expected:  lightstep.Thresholds{},
			shouldErr: true,
		},
		// invalid critical, valid warning
		{
			thresholds: map[string]interface{}{
				"critical": 5,
				"warning":  "10",
			},
			expected:  lightstep.Thresholds{},
			shouldErr: true,
		},
	}

	for _, c := range cases {
		result, err := buildThresholds(c.thresholds)
		if c.shouldErr {
			require.Error(t, err)
		} else {
			require.NoError(t, err)
			require.Equal(t, c.expected.Critical, result.Critical)
			require.Equal(t, c.expected.Warning, result.Warning)
		}
	}
}

func TestBuildLabelFilters(t *testing.T) {
	type filtersCase struct {
		includes []interface{}
		excludes []interface{}
		expected []lightstep.LabelFilter
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
			expected: []lightstep.LabelFilter{
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
			expected: []lightstep.LabelFilter{
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
			expected: []lightstep.LabelFilter{
				includeFilter,
				excludeFilter,
			},
		},
	}

	for _, c := range cases {
		result := buildLabelFilters(c.includes, c.excludes)
		require.Equal(t, c.expected, result)
	}
}

func TestBuildAlertingRules(t *testing.T) {
	type alertingRuleCase struct {
		rules    []interface{}
		expected []lightstep.AlertingRule
	}

	id := "abc123"
	renotify := "1h"
	renotifyMillis := 3600000

	cases := []alertingRuleCase{
		// without includes or excludes
		{
			rules: []interface{}{
				map[string]interface{}{
					"id":       id,
					"renotify": renotify,
				},
			},
			expected: []lightstep.AlertingRule{
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
					"id":       id,
					"renotify": renotify,
					"include_filters": []interface{}{
						map[string]interface{}{
							"key":   k,
							"value": v,
						},
					},
				},
			},
			expected: []lightstep.AlertingRule{
				{
					MessageDestinationID: id,
					UpdateInterval:       renotifyMillis,
					MatchOn:              lightstep.MatchOn{GroupBy: []lightstep.LabelFilter{includeFilter}},
				},
			},
		},
		// with excludes
		{
			rules: []interface{}{
				map[string]interface{}{
					"id":       id,
					"renotify": renotify,
					"exclude_filters": []interface{}{
						map[string]interface{}{
							"key":   k,
							"value": v,
						},
					},
				},
			},
			expected: []lightstep.AlertingRule{
				{
					MessageDestinationID: id,
					UpdateInterval:       renotifyMillis,
					MatchOn:              lightstep.MatchOn{GroupBy: []lightstep.LabelFilter{excludeFilter}},
				},
			},
		},
		// with both includes excludes
		{
			rules: []interface{}{
				map[string]interface{}{
					"id":       id,
					"renotify": renotify,
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
			expected: []lightstep.AlertingRule{
				{
					MessageDestinationID: id,
					UpdateInterval:       renotifyMillis,
					MatchOn:              lightstep.MatchOn{GroupBy: []lightstep.LabelFilter{includeFilter, excludeFilter}},
				},
			},
		},
	}

	for _, c := range cases {
		result, err := buildAlertingRules(c.rules)
		require.NoError(t, err)
		require.Equal(t, c.expected, result)
		require.Equal(t, c.expected, result)
	}
}

func TestBuildGroupBy(t *testing.T) {
	method := "method"
	environment := "kube_environment"
	aggregation := "max"

	expected := lightstep.GroupBy{
		LabelKeys:   []string{method, environment},
		Aggregation: aggregation,
	}

	result := buildGroupBy(aggregation, []interface{}{method, environment})
	require.Equal(t, expected, result)
}

func TestValidateThresholds(t *testing.T) {
	type thresholdCase struct {
		thresholds map[string]interface{}
		shouldErr  bool
	}

	cases := []thresholdCase{
		// valid critical, warning and operand
		{
			thresholds: map[string]interface{}{
				"critical": "2",
				"warning":  "1",
				"operand":  "above",
			},
			shouldErr: false,
		},
		// valid critical and operand
		{
			thresholds: map[string]interface{}{
				"critical": "2",
				"operand":  "above",
			},
			shouldErr: false,
		},
		// missing critical
		{
			thresholds: map[string]interface{}{
				"warning": "1",
				"operand": "above",
			},
			shouldErr: true,
		},
		// incorrect operand 'below'
		{
			thresholds: map[string]interface{}{
				"critical": "2",
				"warning":  "1",
				"operand":  "below",
			},
			shouldErr: true,
		},
		// incorrect operand 'above'
		{
			thresholds: map[string]interface{}{
				"critical": "1",
				"warning":  "2",
				"operand":  "above",
			},
			shouldErr: true,
		},
	}

	for _, c := range cases {
		_, errs := validateThresholds(c.thresholds, "")
		if c.shouldErr {
			require.NotEmpty(t, errs)
		} else {
			require.Empty(t, errs)
		}
	}
}

func TestValidateFilters(t *testing.T) {
	type filterCase struct {
		filters   []interface{}
		expectErr bool
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
			expectErr: false,
		},
		// missing value
		{
			filters: []interface{}{
				map[string]interface{}{
					"key": "key1",
				},
			},
			expectErr: true,
		},
		// missing key
		{
			filters: []interface{}{
				map[string]interface{}{
					"value": "value1",
				},
			},
			expectErr: true,
		},
		// key is not a string
		{
			filters: []interface{}{
				map[string]interface{}{
					"key":   1,
					"value": "some-value",
				},
			},
			expectErr: true,
		},
		// value is not a string
		{
			filters: []interface{}{
				map[string]interface{}{
					"key":   "some-key",
					"value": 1,
				},
			},
			expectErr: true,
		},
	}

	for _, c := range cases {
		err := validateFilters(c.filters)
		if c.expectErr {
			require.Error(t, err)
		} else {
			require.NoError(t, err)
		}
	}
}
