package main

import (
	"fmt"
	"strings"

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
		Importer: &schema.ResourceImporter{
			State: resourceMetricConditionImport,
		},
		Schema: map[string]*schema.Schema{
			"project_name": {
				Type:     schema.TypeString,
				Required: true,
			},
			"name": {
				Type:     schema.TypeString,
				Required: true,
			},
			"type": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"description": {
				Type:     schema.TypeString,
				Optional: true,
			},
			"expression": {
				Type:     schema.TypeList,
				Required: true,
				MaxItems: 1,
				MinItems: 1,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"evaluation_window": {
							Type:         schema.TypeString,
							Required:     true,
							ValidateFunc: validation.StringInSlice(GetValidEvaluationWindows(), false),
						},
						"evaluation_criteria": {
							Type:         schema.TypeString,
							Required:     true,
							ValidateFunc: validation.StringInSlice([]string{"on_average", "at_least_once", "always", "in_total"}, false),
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
						"operand": {
							Type:         schema.TypeString,
							Required:     true,
							ValidateFunc: validation.StringInSlice([]string{"above", "below"}, false),
						},
						"num_sec_per_point": {
							Type:     schema.TypeInt,
							Optional: true,
						},
						"thresholds": {
							Type:     schema.TypeList,
							MaxItems: 1,
							MinItems: 1,
							Elem: &schema.Resource{
								Schema: getThresholdSchema(),
							},
							Required: true,
						},
					},
				},
			},
			"metric_query": {
				Type:     schema.TypeList,
				Required: true,
				Elem: &schema.Resource{
					Schema: getQuerySchema(),
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
		"update_interval": {
			Type:         schema.TypeString,
			Required:     true,
			ValidateFunc: validation.StringInSlice(GetValidUpdateInterval(), false),
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
		"metric": {
			Type:     schema.TypeString,
			Optional: true, // optional for composite formula
		},
		"hidden": {
			Type:     schema.TypeBool,
			Required: true,
		},
		"query_name": {
			Type:     schema.TypeString,
			Required: true,
		},
		"timeseries_operator": {
			Type:         schema.TypeString,
			Optional:     true,
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
			Type:     schema.TypeList,
			MaxItems: 1,
			Elem: &schema.Resource{
				Schema: map[string]*schema.Schema{
					"aggregation_method": {
						Type:         schema.TypeString,
						Optional:     true,
						ValidateFunc: validation.StringInSlice([]string{"sum", "avg", "max", "min", "count", "count_non_zero"}, true),
					},
					"keys": {
						Type: schema.TypeList,
						Elem: &schema.Schema{
							Type: schema.TypeString,
						},
						Optional: true,
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
			Type:     schema.TypeFloat,
			Optional: true,
		},
		"warning": {
			Type:     schema.TypeFloat,
			Optional: true,
		},
	}
}

func resourceMetricConditionCreate(d *schema.ResourceData, m interface{}) error {
	client := m.(*lightstep.Client)

	attributes, err := getMetricConditionAttributesFromResource(d)
	if err != nil {
		return err
	}

	condition := lightstep.MetricCondition{
		Type:       "metric_alert",
		Attributes: *attributes,
	}

	created, err := client.CreateMetricCondition(d.Get("project_name").(string), condition)
	if err != nil {
		return err
	}

	d.SetId(created.ID)
	return resourceMetricConditionRead(d, m)
}

func getMetricConditionAttributesFromResource(d *schema.ResourceData) (*lightstep.MetricConditionAttributes, error) {
	expression := d.Get("expression").([]interface{})[0].(map[string]interface{})

	thresholds := buildThresholds(d)

	attributes := &lightstep.MetricConditionAttributes{
		Type:        "metrics",
		Name:        d.Get("name").(string),
		Description: d.Get("description").(string),
		Expression: lightstep.Expression{
			EvaluationCriteria: expression["evaluation_criteria"].(string),
			IsMulti:            expression["is_multi"].(bool),
			IsNoData:           expression["is_no_data"].(bool),
			Operand:            expression["operand"].(string),
			EvaluationWindow:   validEvaluationWindow[expression["evaluation_window"].(string)],
			Thresholds:         thresholds,
		},
	}

	numSecPerPoint, ok := d.GetOk("expression.0.num-sec-per-point")
	if ok {
		i := numSecPerPoint.(int)
		attributes.Expression.NumSecPerPoint = &i
	}

	queries, err := buildQueries(d.Get("metric_query").([]interface{}))
	if err != nil {
		return nil, err
	}

	attributes.Queries = queries

	alertingRules, err := buildAlertingRules(d.Get("alerting_rule").([]interface{}))
	if err != nil {
		return nil, err
	}

	attributes.AlertingRules = alertingRules
	return attributes, nil
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

	attrs, err := getMetricConditionAttributesFromResource(d)
	if err != nil {
		return err
	}

	_, err = client.UpdateMetricCondition(
		d.Get("project_name").(string),
		d.Id(),
		*attrs,
	)

	return err
}

func resourceMetricConditionDelete(d *schema.ResourceData, m interface{}) error {
	client := m.(*lightstep.Client)
	err := client.DeleteMetricCondition(d.Get("project_name").(string), d.Id())
	return err
}

func buildAlertingRules(alertingRulesIn []interface{}) ([]lightstep.AlertingRule, error) {
	var newRules []lightstep.AlertingRule

	var alertingRules []map[string]interface{}
	for _, ruleIn := range alertingRulesIn {
		alertingRules = append(alertingRules, ruleIn.(map[string]interface{}))
	}

	for _, rule := range alertingRules {
		newRule := lightstep.AlertingRule{
			MessageDestinationID: rule["id"].(string),
		}

		newRule.UpdateInterval = validUpdateInterval[rule["update_interval"].(string)]

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

func buildQueries(queriesIn []interface{}) ([]lightstep.MetricQueryWithAttributes, error) {
	var newQueries []lightstep.MetricQueryWithAttributes
	var queries []map[string]interface{}
	for _, queryIn := range queriesIn {
		queries = append(queries, queryIn.(map[string]interface{}))
	}

	for _, query := range queries {
		metric := query["metric"].(string)
		queryType := "single"
		if metric == "" {
			queryType = "composite"
		}
		newQuery := lightstep.MetricQueryWithAttributes{
			Name:    query["query_name"].(string),
			Type:    queryType,
			Hidden:  query["hidden"].(bool),
			Display: "line",
			Query: lightstep.MetricQuery{
				TimeseriesOperator: query["timeseries_operator"].(string),
				Metric:             metric,
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

		groupBy := query["group_by"]
		err := validateGroupBy(groupBy, newQuery.Type)
		if err != nil {
			return nil, err
		}

		if newQuery.Type == "single" {
			arr := groupBy.([]interface{})
			g := arr[0].(map[string]interface{})

			newQuery.Query.GroupBy =
				lightstep.GroupBy{
					Aggregation: g["aggregation_method"].(string),
					LabelKeys:   buildKeys(g["keys"].([]interface{})),
				}
		}
		newQueries = append(newQueries, newQuery)
	}
	return newQueries, nil
}

func buildKeys(keysIn []interface{}) []string {
	var keys []string
	for _, k := range keysIn {
		keys = append(keys, k.(string))
	}
	return keys
}

func buildThresholds(d *schema.ResourceData) lightstep.Thresholds {
	t := lightstep.Thresholds{}

	critical, cOk := d.GetOk("expression.0.thresholds.0.critical")
	if cOk {
		c := critical.(float64)
		t.Critical = &c
	}

	warning, wOk := d.GetOk("expression.0.thresholds.0.warning")
	if wOk {
		w := warning.(float64)
		t.Warning = &w
	}

	return t
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

func validateGroupBy(groupBy interface{}, queryType string) error {
	// groupBy is invalid for composite queries
	if queryType == "composite" {
		if groupBy != nil && len(groupBy.([]interface{})) > 0 {
			return fmt.Errorf("invalid block group_by found on composite")
		}
		return nil
	}

	if groupBy == nil || len(groupBy.([]interface{})) == 0 {
		return fmt.Errorf("missing fields in group_by found on composite")
	}

	g := groupBy.([]interface{})[0].(map[string]interface{})
	_, hasAggMethod := g["aggregation_method"]
	if !hasAggMethod {
		return fmt.Errorf("missing required field aggregation_method on group_by")
	}

	return nil
}

func resourceMetricConditionImport(d *schema.ResourceData, m interface{}) ([]*schema.ResourceData, error) {
	client := m.(*lightstep.Client)

	ids := strings.Split(d.Id(), ".")
	if len(ids) != 2 {
		return []*schema.ResourceData{}, fmt.Errorf("Error importing lightstep_metric_condition. Expecting an  ID formed as '<lightstep_project>.<lightstep_metric_condition_ID>'")
	}

	project, id := ids[0], ids[1]
	c, err := client.GetMetricCondition(project, id)
	if err != nil {
		return []*schema.ResourceData{}, err
	}

	d.SetId(id)

	err = d.Set("project_name", project)
	if err != nil {
		return nil, err
	}

	err = d.Set("name", c.Attributes.Name)
	if err != nil {
		return nil, err
	}

	err = d.Set("description", c.Attributes.Description)
	if err != nil {
		return nil, err
	}

	err = d.Set("type", "metric_alert")
	if err != nil {
		return nil, err
	}

	thresholdEntries := map[string]interface{}{}

	if c.Attributes.Expression.Thresholds.Critical != nil {
		thresholdEntries["critical"] = *c.Attributes.Expression.Thresholds.Critical
	}
	if c.Attributes.Expression.Thresholds.Warning != nil {
		thresholdEntries["warning"] = *c.Attributes.Expression.Thresholds.Warning
	}

	err = d.Set("expression", []map[string]interface{}{
		{
			"evaluation_window":   GetEvaluationWindowValue(c.Attributes.Expression.EvaluationWindow),
			"evaluation_criteria": c.Attributes.Expression.EvaluationCriteria,
			"is_multi":            c.Attributes.Expression.IsMulti,
			"is_no_data":          c.Attributes.Expression.IsNoData,
			"operand":             c.Attributes.Expression.Operand,
			"num-sec-per-point":   c.Attributes.Expression.NumSecPerPoint,
			"thresholds": []interface{}{
				thresholdEntries,
			},
		},
	})
	if err != nil {
		return nil, err
	}

	var queries []interface{}
	for _, q := range c.Attributes.Queries {
		includeFilters, excludeFilters := getIncludeExcludeFilters(q.Query.Filters)

		var groupBy []interface{}
		if len(q.Query.GroupBy.LabelKeys) > 0 {
			groupBy = []interface{}{
				map[string]interface{}{
					"aggregation_method": q.Query.GroupBy.Aggregation,
					"keys":               q.Query.GroupBy.LabelKeys,
				},
			}
		}
		queries = append(queries, map[string]interface{}{
			"metric":              q.Query.Metric,
			"hidden":              q.Hidden,
			"query_name":          q.Name,
			"timeseries_operator": q.Query.TimeseriesOperator,
			"include_filters":     includeFilters,
			"exclude_filters":     excludeFilters,
			"group_by":            groupBy,
		})
	}

	err = d.Set("metric_query", queries)
	if err != nil {
		return nil, err
	}

	var alertingRules []interface{}
	for _, r := range c.Attributes.AlertingRules {
		includeFilters, excludeFilters := getIncludeExcludeFilters(r.MatchOn.GroupBy)

		alertingRules = append(alertingRules, map[string]interface{}{
			"id":              r.MessageDestinationID,
			"update_interval": GetUpdateIntervalValue(r.UpdateInterval),
			"include_filters": includeFilters,
			"exclude_filters": excludeFilters,
		})
	}
	err = d.Set("alerting_rule", alertingRules)
	if err != nil {
		return nil, err
	}

	return []*schema.ResourceData{d}, nil
}

func getIncludeExcludeFilters(filters []lightstep.LabelFilter) ([]interface{}, []interface{}) {
	var includeFilters []interface{}
	var excludeFilters []interface{}
	for _, f := range filters {
		if f.Operand == "eq" {
			includeFilters = append(includeFilters, map[string]interface{}{
				"key":   f.Key,
				"value": f.Value,
			})
		} else {
			excludeFilters = append(excludeFilters, map[string]interface{}{
				"key":   f.Key,
				"value": f.Value,
			})
		}
	}
	return includeFilters, excludeFilters
}
