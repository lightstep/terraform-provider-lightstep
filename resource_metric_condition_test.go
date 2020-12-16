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

  metric_query {
    metric         = "requests"
    query_name          = "a"
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
      aggregation_method = "avg"
      keys = ["method"]
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

	updatedConditionConfig := `
resource "lightstep_slack_destination" "slack" {
  project_name = "terraform-provider-tests"
  channel = "#emergency-room"
}

resource "lightstep_metric_condition" "test" {
  project_name = "terraform-provider-tests"
  name = "updated"

  expression {
	  evaluation_window   = "1h" 
	  evaluation_criteria = "at_least_once"
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
      aggregation_method = "avg"
      keys = ["method"]
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
				ExpectError: regexp.MustCompile("required field is not set"),
			},
			{
				Config: conditionConfig,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckMetricConditionExists(resourceName, &condition),
					resource.TestCheckResourceAttr(resourceName, "name", "Too many requests"),
					// TODO: verify more fields here, I don't understand how to do nested fields
				),
			},
			{
				Config: updatedConditionConfig,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckMetricConditionExists(resourceName, &condition),
					resource.TestCheckResourceAttr(resourceName, "name", "updated"),
					// TODO: verify more fields here, I don't understand how to do nested fields
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
	}

	cases := []thresholdCase{
		// valid critical
		{
			thresholds: map[string]interface{}{
				"critical": 5.0,
			},
			expected: lightstep.Thresholds{
				Critical: 5,
			},
		},
		// valid critical and warning
		{
			thresholds: map[string]interface{}{
				"critical": 5.0,
				"warning":  10.0,
			},
			expected: lightstep.Thresholds{
				Critical: 5.0,
				Warning:  10.0,
			},
		},
	}

	for _, c := range cases {
		result := buildThresholds(c.thresholds)
		require.Equal(t, c.expected.Critical, result.Critical)
		require.Equal(t, c.expected.Warning, result.Warning)
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
					"id":              id,
					"update_interval": renotify,
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
			expectErr: true,
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
