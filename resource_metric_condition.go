package main

import (
	"fmt"
	"strconv"

	"github.com/hashicorp/terraform-plugin-sdk/helper/validation"

	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
	"github.com/lightstep/terraform-provider-lightstep/lightstep"
)

func resourceMetricCondition() *schema.Resource {
	return &schema.Resource{
		Create: resourceMetricConditionCreate,
		Read:   resourceMetricConditionRead,
		Update: resourceMetricConditionUpdate,
		Delete: resourceMetricConditionDelete,
		Schema: map[string]*schema.Schema{
			"project_name": {
				Type:     schema.TypeString,
				Required: true,
			},
			"condition_name": {
				Type:     schema.TypeString,
				Required: true,
			},
			"type": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"thresholds": {
				Type: schema.TypeMap,
				Elem: &schema.Resource{
					Schema: getThresholdSchema(),
				},
				Required:     true,
				ValidateFunc: validateThresholds,
			},
			"evaluation_window": {
				Type:         schema.TypeString,
				Required:     true,
				ValidateFunc: validation.StringInSlice(validEvaluationWindowInput, false),
			},
			"evaluation_criteria": {
				Type:         schema.TypeString,
				Required:     true,
				ValidateFunc: validation.StringInSlice([]string{"on_average", "at_least_once", "always", "in_total"}, false),
			},
			"display": {
				Type:         schema.TypeString,
				Required:     true,
				ValidateFunc: validation.StringInSlice([]string{"line", "bar", "area"}, false),
			},
			"is_multi": {
				Type:     schema.TypeBool,
				Optional: true,
				Default:  false,
			},
			"is_no_data": {
				Type:     schema.TypeBool,
				Optional: true,
				Default:  false,
			},
			"query": {
				Type:     schema.TypeList,
				Required: true,
				Elem: &schema.Resource{
					Schema: getQuerySchema(),
				},
			},
			"composite_query": {
				Type:     schema.TypeMap,
				Optional: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"query_formula": {
							Type:     schema.TypeString,
							Required: true,
						},
					},
				},
			},
			"alerting_rule": {
				Type:     schema.TypeList,
				Optional: true,
				Elem: &schema.Resource{
					Schema: getAlertingRuleSchema(),
				},
			},
		},
	}
}

func getAlertingRuleSchema() map[string]*schema.Schema {
	return map[string]*schema.Schema{
		"renotify": {
			Type:         schema.TypeString,
			Required:     true,
			ValidateFunc: validation.StringInSlice(validRenotifyInput, false),
		},
		"id": {
			Type:     schema.TypeString,
			Required: true,
		},
		"include_filters": {
			Type:     schema.TypeList,
			Elem:     &schema.Schema{Type: schema.TypeMap},
			Optional: true,
		},
		"exclude_filters": {
			Type:     schema.TypeList,
			Elem:     &schema.Schema{Type: schema.TypeMap},
			Optional: true,
		},
	}
}

func getQuerySchema() map[string]*schema.Schema {
	return map[string]*schema.Schema{
		"metric_name": {
			Type:     schema.TypeString,
			Required: true,
		},
		"hidden": {
			Type:     schema.TypeBool,
			Required: true,
		},
		"query_name": {
			Type:         schema.TypeString,
			Required:     true,
			ValidateFunc: validation.StringInSlice(validQueryNames, false),
		},
		"timeseries_operator": {
			Type:         schema.TypeString,
			Required:     true,
			ValidateFunc: validation.StringInSlice([]string{"rate", "delta", "mean", "last"}, false),
		},
		"include_filters": {
			Type:     schema.TypeList,
			Elem:     &schema.Schema{Type: schema.TypeMap},
			Optional: true,
		},
		"exclude_filters": {
			Type:     schema.TypeList,
			Elem:     &schema.Schema{Type: schema.TypeMap},
			Optional: true,
		},
		"group_by": {
			Type: schema.TypeSet,
			Elem: &schema.Resource{
				Schema: map[string]*schema.Schema{
					"aggregation": {
						Type:         schema.TypeString,
						Required:     true,
						ValidateFunc: validation.StringInSlice([]string{"sum", "avg", "max", "min", "count", "count_non_zero"}, true),
					},
					"keys": {
						Type:     schema.TypeList,
						Elem:     &schema.Schema{Type: schema.TypeString},
						Required: true,
					},
				},
			},
			Optional: true,
		},
	}
}

func getThresholdSchema() map[string]*schema.Schema {
	return map[string]*schema.Schema{
		"critical": {
			Type:     schema.TypeInt,
			Required: true,
		},
		"warning": {
			Type:     schema.TypeInt,
			Optional: true,
		},
		"operand": {
			Type:         schema.TypeString,
			Required:     true,
			ValidateFunc: validation.StringInSlice([]string{"above", "below"}, false),
		},
	}
}

func resourceMetricConditionCreate(d *schema.ResourceData, m interface{}) error {
	client := m.(*lightstep.Client)

	condition := lightstep.MetricCondition{
		Type: "metric_alert",
	}

	condition.Attributes = lightstep.MetricConditionAttributes{
		Type: "metrics",
		Name: d.Get("condition_name").(string),
		Expression: lightstep.Expression{
			EvaluationCriteria: d.Get("evaluation_criteria").(string),
			IsMulti:            d.Get("is_multi").(bool),
			IsNoData:           d.Get("is_no_data").(bool),
		},
	}

	evaluationWindowStr := d.Get("evaluation_window").(string)
	condition.Attributes.EvaluationWindow = validEvaluationWindow[evaluationWindowStr]

	tfThresholds := d.Get("thresholds").(map[string]interface{})
	thresholds, err := buildThresholds(tfThresholds)
	if err != nil {
		return err
	}
	condition.Attributes.Thresholds = thresholds
	condition.Attributes.Operand = tfThresholds["operand"].(string)

	display := d.Get("display").(string)

	queries, err := buildSingleQueries(d.Get("query").([]interface{}), display)
	if err != nil {
		return err
	}

	compositeQuery, found := d.GetOk("composite_query")
	if found {
		c := compositeQuery.(map[string]interface{})
		query := lightstep.MetricQueryWithAttributes{
			Name:    c["query_formula"].(string),
			Type:    "composite",
			Hidden:  false,
			Display: display,
			Query:   lightstep.MetricQuery{},
		}
		queries = append(queries, query)
	}

	condition.Attributes.Queries = queries

	alertingRules, err := buildAlertingRules(d.Get("alerting_rule").([]interface{}))
	if err != nil {
		return err
	}

	condition.Attributes.AlertingRules = alertingRules

	created, err := client.CreateMetricCondition(d.Get("project_name").(string), condition)
	if err != nil {
		return err
	}

	d.SetId(created.ID)
	return resourceMetricConditionRead(d, m)
}

func resourceMetricConditionRead(d *schema.ResourceData, m interface{}) error {
	client := m.(*lightstep.Client)

	_, err := client.GetMetricCondition(d.Get("project_name").(string), d.Id())
	if err != nil {
		return err
	}
	return nil
}

func resourceMetricConditionUpdate(d *schema.ResourceData, m interface{}) error {
	client := m.(*lightstep.Client)

	attrs := lightstep.MetricConditionAttributes{
		Type: "metrics",
		Name: d.Get("condition_name").(string),
		Expression: lightstep.Expression{
			EvaluationCriteria: d.Get("evaluation_criteria").(string),
			IsMulti:            d.Get("is_multi").(bool),
			IsNoData:           d.Get("is_no_data").(bool),
		},
	}

	evaluationWindowStr := d.Get("evaluation_window").(string)
	attrs.EvaluationWindow = validEvaluationWindow[evaluationWindowStr]

	tfThresholds := d.Get("thresholds").(map[string]interface{})
	thresholds, err := buildThresholds(tfThresholds)
	if err != nil {
		return err
	}
	attrs.Thresholds = thresholds
	attrs.Operand = tfThresholds["operand"].(string)

	display := d.Get("display").(string)

	queries, err := buildSingleQueries(d.Get("query").([]interface{}), display)
	if err != nil {
		return err
	}

	compositeQuery, found := d.GetOk("composite_query")
	if found {
		c := compositeQuery.(map[string]interface{})
		query := lightstep.MetricQueryWithAttributes{
			Name:    c["query_formula"].(string),
			Type:    "composite",
			Hidden:  false,
			Display: display,
			Query:   lightstep.MetricQuery{},
		}
		queries = append(queries, query)
	}

	attrs.Queries = queries

	alertingRules, err := buildAlertingRules(d.Get("alerting_rule").([]interface{}))
	if err != nil {
		return err
	}

	attrs.AlertingRules = alertingRules

	_, err = client.UpdateMetricCondition(
		d.Get("project_name").(string),
		d.Id(),
		attrs,
	)

	return err
}

func resourceMetricConditionDelete(d *schema.ResourceData, m interface{}) error {
	client := m.(*lightstep.Client)
	err := client.DeleteMetricCondition(d.Get("project_name").(string), d.Id())
	return err
}

func buildAlertingRules(alertingRulesIn []interface{}) ([]lightstep.AlertingRule, error) {
	newRules := []lightstep.AlertingRule{}

	alertingRules := []map[string]interface{}{}
	for _, ruleIn := range alertingRulesIn {
		alertingRules = append(alertingRules, ruleIn.(map[string]interface{}))
	}

	for _, rule := range alertingRules {
		newRule := lightstep.AlertingRule{
			MessageDestinationID: rule["id"].(string),
		}

		newRule.UpdateInterval = validRenotify[rule["renotify"].(string)]

		var includes []interface{}
		var excludes []interface{}

		filters := rule["include_filters"]
		if filters != nil {
			err := validateFilters(filters.([]interface{}))
			if err != nil {
				return nil, err
			}
			includes = filters.([]interface{})
		}

		filters = rule["exclude_filters"]
		if filters != nil {
			err := validateFilters(filters.([]interface{}))
			if err != nil {
				return nil, err
			}
			excludes = filters.([]interface{})
		}

		newFilters := buildLabelFilters(includes, excludes)
		newRule.MatchOn = lightstep.MatchOn{GroupBy: newFilters}

		newRules = append(newRules, newRule)
	}
	return newRules, nil
}

func buildSingleQueries(queriesIn []interface{}, display string) ([]lightstep.MetricQueryWithAttributes, error) {
	newQueries := []lightstep.MetricQueryWithAttributes{}
	queries := []map[string]interface{}{}
	for _, queryIn := range queriesIn {
		queries = append(queries, queryIn.(map[string]interface{}))
	}

	for _, query := range queries {
		newQuery := lightstep.MetricQueryWithAttributes{
			Name:    query["query_name"].(string),
			Type:    "single",
			Hidden:  query["hidden"].(bool),
			Display: display,
			Query: lightstep.MetricQuery{
				Metric:             query["metric_name"].(string),
				TimeseriesOperator: query["timeseries_operator"].(string),
			},
		}

		includes := query["include_filters"]
		if includes != nil {
			err := validateFilters(includes.([]interface{}))
			if err != nil {
				return nil, err
			}
		}

		excludes := query["exclude_filters"]
		if excludes != nil {
			err := validateFilters(excludes.([]interface{}))
			if err != nil {
				return nil, err
			}
		}

		filters := buildLabelFilters(includes.([]interface{}), excludes.([]interface{}))
		newQuery.Query.Filters = filters

		groupBy, ok := query["group_by"]
		if ok {
			g := groupBy.(*schema.Set).List()[0].(map[string]interface{})
			newQuery.Query.GroupBy = buildGroupBy(g["aggregation"].(string), g["keys"].([]interface{}))
		}
		newQueries = append(newQueries, newQuery)
	}
	return newQueries, nil
}

func buildGroupBy(aggregation string, labelKeys []interface{}) lightstep.GroupBy {
	groupBy := lightstep.GroupBy{
		Aggregation: aggregation,
	}

	keys := []string{}
	for _, k := range labelKeys {
		keys = append(keys, k.(string))
		groupBy.LabelKeys = keys
	}
	return groupBy
}

func buildThresholds(thresholds map[string]interface{}) (lightstep.Thresholds, error) {
	t := lightstep.Thresholds{}

	critical, ok := thresholds["critical"]
	if ok {
		critical, isString := critical.(string)
		if !isString {
			return t, fmt.Errorf("thresholds must be string, got: %T", critical)
		}

		c, err := strconv.Atoi(critical)
		if err != nil {
			return t, err
		}
		t.Critical = c
	}

	warning, ok := thresholds["warning"]
	if ok {
		warning, isString := warning.(string)
		if !isString {
			return t, fmt.Errorf("thresholds must be string, got: %T", warning)
		}
		w, err := strconv.Atoi(warning)
		if err != nil {
			return t, err
		}
		t.Warning = w
	}
	return t, nil
}

func buildLabelFilters(includes []interface{}, excludes []interface{}) []lightstep.LabelFilter {
	var filters []lightstep.LabelFilter

	if len(includes) > 0 {
		for _, includeFilter := range includes {
			key := includeFilter.(map[string]interface{})["key"]
			value := includeFilter.(map[string]interface{})["value"]

			filters = append(filters, lightstep.LabelFilter{
				Operand: "eq",
				Key:     key.(string),
				Value:   value.(string),
			})
		}
	}

	if len(excludes) > 0 {
		for _, excludeFilter := range excludes {
			key := excludeFilter.(map[string]interface{})["key"]
			value := excludeFilter.(map[string]interface{})["value"]
			filters = append(filters, lightstep.LabelFilter{
				Operand: "neq",
				Key:     key.(string),
				Value:   value.(string),
			})
		}
	}
	return filters
}

func validateThresholds(val interface{}, _ string) (warns []string, errors []error) {
	value := val.(map[string]interface{})

	criticalStr, ok := value["critical"].(string)
	if !ok {
		return nil, []error{fmt.Errorf("missing critical threshold")}
	}

	critical, err := strconv.Atoi(criticalStr)
	if err != nil {
		return nil, []error{fmt.Errorf("invalid threshold: %v", err)}
	}

	warningStr, ok := value["warning"].(string)
	var warning int
	if !ok {
		return
	}

	warning, err = strconv.Atoi(warningStr)
	if err != nil {
		return nil, []error{fmt.Errorf("invalid threshold: %v", err)}
	}

	operand := value["operand"].(string)

	switch operand {
	case "above":
		if warning > critical {
			errors = append(errors, fmt.Errorf("warning cannot be above critical with operand %s", operand))
		}
	case "below":
		if warning < critical {
			errors = append(errors, fmt.Errorf("warning cannot be below critical with operand %s", operand))
		}
	}
	return
}

func validateFilters(filters []interface{}) error {
	for _, filter := range filters {
		key, ok := filter.(map[string]interface{})["key"]
		if !ok {
			return fmt.Errorf("'key' is a required field")
		}

		value, ok := filter.(map[string]interface{})["value"]
		if !ok {
			return fmt.Errorf("'value' is a required field")
		}

		switch key.(type) {
		case string:
		default:
			return fmt.Errorf("value must be a string. got: %v", key)
		}

		switch value.(type) {
		case string:
		default:
			return fmt.Errorf("value must be a string. got: %T", value)
		}
	}
	return nil
}
