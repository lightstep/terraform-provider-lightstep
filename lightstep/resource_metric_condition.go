package lightstep

import (
	"context"
	"fmt"
	"net/http"
	"strconv"
	"strings"

	"github.com/hashicorp/go-cty/cty"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"

	"github.com/lightstep/terraform-provider-lightstep/client"
)

// resourceUnifiedCondition creates a resource for either:
//
// (1) The legacy lightstep_metric_condition
// (2) The unified lightstep_condition
//
// The resources are largely the same with the primary difference being the
// query format.
func resourceUnifiedCondition(conditionSchemaType ConditionSchemaType) *schema.Resource {
	p := resourceUnifiedConditionImp{conditionSchemaType: conditionSchemaType}

	resource := &schema.Resource{
		CreateContext: p.resourceUnifiedConditionCreate,
		ReadContext:   p.resourceUnifiedConditionRead,
		UpdateContext: p.resourceUnifiedConditionUpdate,
		DeleteContext: p.resourceUnifiedConditionDelete,
		Importer: &schema.ResourceImporter{
			StateContext: p.resourceUnifiedConditionImport,
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
						"thresholds": {
							Type:     schema.TypeList,
							Required: true,
							MaxItems: 1,
							MinItems: 1,
							Elem: &schema.Resource{
								Schema: getThresholdSchema(),
							},
						},
					},
				},
			},
			"alerting_rule": {
				Type:     schema.TypeSet,
				Optional: true,
				Elem: &schema.Resource{
					Schema: getAlertingRuleSchema(),
				},
			},
		},
	}

	if conditionSchemaType == UnifiedConditionSchema {
		resource.Schema["query"] = &schema.Schema{
			Type:     schema.TypeList,
			Required: true,
			Elem: &schema.Resource{
				Schema: getUnifiedQuerySchema(),
			},
		}
	} else {
		resource.Schema["metric_query"] = &schema.Schema{
			Type:     schema.TypeList,
			Required: true,
			Elem: &schema.Resource{
				Schema: getMetricQuerySchema(),
			},
		}
	}
	return resource
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
		"filters": {
			Type:        schema.TypeList,
			Elem:        &schema.Schema{Type: schema.TypeMap},
			Description: "Non-equality filters (operand: contains, regexp, etc)",
			Optional:    true,
		},
	}
}

func getSpansQuerySchema() *schema.Schema {
	sma := schema.Schema{
		Type:     schema.TypeList,
		MaxItems: 1,
		Elem: &schema.Resource{
			Schema: map[string]*schema.Schema{
				"query": {
					Type:     schema.TypeString,
					Required: true,
				},
				"operator": {
					Type:         schema.TypeString,
					ValidateFunc: validation.StringInSlice([]string{"latency", "rate", "error_ratio", "count"}, false),
					Required:     true,
				},
				"operator_input_window_ms": {
					Type:     schema.TypeInt,
					Optional: true,
					// Duration micros must be at least 30s and an even number of seconds
					ValidateFunc: validation.All(validation.IntDivisibleBy(1_000), validation.IntAtLeast(30_000)),
				},
				"group_by_keys": {
					Type:     schema.TypeList,
					Optional: true,
					Elem: &schema.Schema{
						Type: schema.TypeString,
					},
				},
				"latency_percentiles": {
					Type:     schema.TypeList,
					Optional: true,
					Computed: true,
					Elem: &schema.Schema{
						ValidateDiagFunc: func(l interface{}, p cty.Path) diag.Diagnostics {
							latency, ok := l.(float64)
							var diags diag.Diagnostics
							if ok && (latency < 0 || latency > 100) {
								oneDiag := diag.Diagnostic{
									Severity: diag.Error,
									Summary:  "wrong value",
									Detail:   "latency_percentiles must be between 0 and 100",
								}
								diags = append(diags, oneDiag)
							}
							return diags
						},
						Type: schema.TypeFloat,
					},
				},
			},
		},
		Optional: true,
	}
	return &sma
}

func getMetricQuerySchema() map[string]*schema.Schema {
	sma := map[string]*schema.Schema{
		"metric": {
			Type:     schema.TypeString,
			Optional: true, // optional for composite formula
			Computed: true,
		},
		"hidden": {
			Type:     schema.TypeBool,
			Required: true,
		},
		"display": {
			Type:         schema.TypeString,
			Optional:     true,
			ValidateFunc: validation.StringInSlice([]string{"line", "area", "bar", "big_number", "heatmap", "dependency_map"}, false),
		},
		"query_name": {
			Type:     schema.TypeString,
			Required: true,
		},
		"timeseries_operator": {
			Type:         schema.TypeString,
			Optional:     true,
			Computed:     true,
			ValidateFunc: validation.StringInSlice([]string{"rate", "delta", "last", "min", "max", "avg"}, false),
		},
		"timeseries_operator_input_window_ms": {
			Type:         schema.TypeInt,
			Description:  "Unit specified in milliseconds, but must be at least 30,000 and a round number of seconds (i.e. evenly divisible by 1,000)",
			Optional:     true,
			ValidateFunc: validation.All(validation.IntDivisibleBy(1_000), validation.IntAtLeast(30_000)),
		},
		"final_window_operation": getFinalWindowOperationSchema(),
		"filters": {
			Type:        schema.TypeList,
			Elem:        &schema.Schema{Type: schema.TypeMap},
			Description: "Non-equality filters (operand: contains, regexp)",
			Optional:    true,
			Computed:    true,
		},
		"include_filters": {
			Type:        schema.TypeList,
			Elem:        &schema.Schema{Type: schema.TypeMap},
			Description: "Equality filters (operand: eq)",
			Optional:    true,
			Computed:    true,
		},
		"exclude_filters": {
			Type:        schema.TypeList,
			Elem:        &schema.Schema{Type: schema.TypeMap},
			Description: "Not-equals filters (operand: neq)",
			Optional:    true,
			Computed:    true,
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
		"tql": {
			Deprecated:  "Use lightstep_dashboard or lightstep_condition instead",
			Description: "Deprecated, use the query_string field in lightstep_dashboard or lightstep_condition instead",
			Type:        schema.TypeString,
			Optional:    true,
			Computed:    true,
		},
		"spans": getSpansQuerySchema(),
	}
	return sma
}

func getFinalWindowOperationSchema() *schema.Schema {
	return &schema.Schema{
		Type:     schema.TypeList,
		Optional: true,
		MaxItems: 1,
		MinItems: 1,
		Elem: &schema.Resource{
			Schema: map[string]*schema.Schema{
				"operator": {
					Type:         schema.TypeString,
					Optional:     true,
					ValidateFunc: validation.StringInSlice([]string{"min", "max", "avg"}, false),
				},
				"input_window_ms": {
					Type:         schema.TypeInt,
					Description:  "Unit specified in milliseconds, but must be at least 30,000 and a round number of seconds (i.e. evenly divisible by 1,000)",
					Optional:     true,
					ValidateFunc: validation.All(validation.IntDivisibleBy(1_000), validation.IntAtLeast(30_000)),
				},
			},
		},
	}
}

func getThresholdSchema() map[string]*schema.Schema {
	return map[string]*schema.Schema{
		"critical": {
			Type:     schema.TypeString,
			Optional: true,
			Default:  "",
		},
		"warning": {
			Type:     schema.TypeString,
			Optional: true,
			Default:  "",
		},
	}
}

type ConditionSchemaType int

const (
	MetricConditionSchema ConditionSchemaType = iota
	UnifiedConditionSchema
)

type resourceUnifiedConditionImp struct {
	conditionSchemaType ConditionSchemaType
}

func (p *resourceUnifiedConditionImp) resourceUnifiedConditionCreate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	c := m.(*client.Client)
	attributes, err := getUnifiedConditionAttributesFromResource(d, p.conditionSchemaType)
	if err != nil {
		return diag.FromErr(fmt.Errorf("failed to get metric condition attributes from resource : %v", err))
	}

	condition := client.UnifiedCondition{
		Type:       "metric_alert",
		Attributes: *attributes,
	}

	created, err := c.CreateUnifiedCondition(ctx, d.Get("project_name").(string), condition)
	if err != nil {
		return diag.FromErr(fmt.Errorf("failed to create metric condition: %v", err))
	}

	d.SetId(created.ID)
	return p.resourceUnifiedConditionRead(ctx, d, m)
}

func (p *resourceUnifiedConditionImp) resourceUnifiedConditionRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	var diags diag.Diagnostics

	c := m.(*client.Client)
	cond, err := c.GetUnifiedCondition(ctx, d.Get("project_name").(string), d.Id())
	if err != nil {
		apiErr, ok := err.(client.APIResponseCarrier)
		if !ok {
			return diag.FromErr(fmt.Errorf("failed to get metric condition: %v", err))
		}

		if apiErr.GetStatusCode() == http.StatusNotFound {
			d.SetId("")
			return diags
		}

		return diag.FromErr(fmt.Errorf("failed to get metric condition: %v", apiErr))
	}

	if err := setResourceDataFromUnifiedCondition(d.Get("project_name").(string), *cond, d, p.conditionSchemaType); err != nil {
		return diag.FromErr(fmt.Errorf("failed to set metric condition from API response to terraform state: %v", err))
	}

	return diags
}

func (p *resourceUnifiedConditionImp) resourceUnifiedConditionUpdate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	c := m.(*client.Client)
	attrs, err := getUnifiedConditionAttributesFromResource(d, p.conditionSchemaType)
	if err != nil {
		return diag.FromErr(fmt.Errorf("failed to get metric condition attributes from resource : %v", err))
	}

	if _, err := c.UpdateUnifiedCondition(ctx, d.Get("project_name").(string), d.Id(), *attrs); err != nil {
		return diag.FromErr(fmt.Errorf("failed to update metric condition: %v", err))
	}

	return p.resourceUnifiedConditionRead(ctx, d, m)
}

func (p *resourceUnifiedConditionImp) resourceUnifiedConditionDelete(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	var diags diag.Diagnostics

	c := m.(*client.Client)
	if err := c.DeleteUnifiedCondition(ctx, d.Get("project_name").(string), d.Id()); err != nil {
		return diag.FromErr(fmt.Errorf("failed to delete metrics condition: %v", err))
	}

	// d.SetId("") is automatically called assuming delete returns no errors, but
	// it is added here for explicitness.
	d.SetId("")
	return diags
}

func (p *resourceUnifiedConditionImp) resourceUnifiedConditionImport(ctx context.Context, d *schema.ResourceData, m interface{}) ([]*schema.ResourceData, error) {
	clnt := m.(*client.Client)

	ids := strings.Split(d.Id(), ".")
	if len(ids) != 2 {
		resourceName := "lighstep_condition"
		if p.conditionSchemaType == MetricConditionSchema {
			resourceName = "lightstep_metric_condition"
		}
		return []*schema.ResourceData{}, fmt.Errorf("error importing %v. Expecting an  ID formed as '<lightstep_project>.<%v_ID>'", resourceName, resourceName)
	}

	project, id := ids[0], ids[1]
	c, err := clnt.GetUnifiedCondition(ctx, project, id)
	if err != nil {
		return []*schema.ResourceData{}, fmt.Errorf("failed to get metric condition. err: %v", err)
	}

	d.SetId(id)
	if err := setResourceDataFromUnifiedCondition(project, *c, d, p.conditionSchemaType); err != nil {
		return nil, fmt.Errorf("failed to set metric condition from API response to terraform state: %v", err)
	}

	return []*schema.ResourceData{d}, nil
}

func getUnifiedConditionAttributesFromResource(d *schema.ResourceData, schemaType ConditionSchemaType) (*client.UnifiedConditionAttributes, error) {
	expression := d.Get("expression").([]interface{})[0].(map[string]interface{})

	thresholds, err := buildThresholds(d)
	if err != nil {
		return nil, err
	}

	attributes := &client.UnifiedConditionAttributes{
		Type:        "metrics",
		Name:        d.Get("name").(string),
		Description: d.Get("description").(string),
		Expression: client.Expression{
			IsMulti:    expression["is_multi"].(bool),
			IsNoData:   expression["is_no_data"].(bool),
			Operand:    expression["operand"].(string),
			Thresholds: thresholds,
		},
	}

	schemaQuery := d.Get("metric_query")
	if schemaType == UnifiedConditionSchema {
		schemaQuery = d.Get("query")
	}
	queries, err := buildQueries(schemaQuery.([]interface{}))
	if err != nil {
		return nil, err
	}

	attributes.Queries = queries

	alertingRules, err := buildAlertingRules(d.Get("alerting_rule").(*schema.Set))
	if err != nil {
		return nil, err
	}

	attributes.AlertingRules = alertingRules
	return attributes, nil
}

func buildAlertingRules(alertingRulesIn *schema.Set) ([]client.AlertingRule, error) {
	var newRules []client.AlertingRule

	var alertingRules []map[string]interface{}
	for _, ruleIn := range alertingRulesIn.List() {
		alertingRules = append(alertingRules, ruleIn.(map[string]interface{}))
	}

	for _, rule := range alertingRules {
		if rule["id"] == "" {
			// This indicates the alerting destination was changed (and as such shows up as a different set entry
			// We just want to skip over these.
			continue
		}

		newRule := client.AlertingRule{
			MessageDestinationID: rule["id"].(string),
		}

		newRule.UpdateInterval = validUpdateInterval[rule["update_interval"].(string)]

		var includes []interface{}
		var excludes []interface{}
		var all []interface{}

		filters := rule["include_filters"]
		if filters != nil {
			err := validateFilters(filters.([]interface{}), false)
			if err != nil {
				return nil, err
			}
			includes = filters.([]interface{})
		}

		filters = rule["exclude_filters"]
		if filters != nil {
			err := validateFilters(filters.([]interface{}), false)
			if err != nil {
				return nil, err
			}
			excludes = filters.([]interface{})
		}

		filters = rule["filters"]
		if filters != nil {
			err := validateFilters(filters.([]interface{}), true)
			if err != nil {
				return nil, err
			}
			all = filters.([]interface{})
		}

		newFilters := buildLabelFilters(includes, excludes, all)
		newRule.MatchOn = client.MatchOn{GroupBy: newFilters}

		newRules = append(newRules, newRule)
	}
	return newRules, nil
}

func buildSpansGroupByKeys(keysIn []interface{}) []string {
	var keys []string
	for _, k := range keysIn {
		keys = append(keys, k.(string))
	}
	return keys
}

func buildLatencyPercentiles(lats []interface{}, display string) []float64 {
	if display == "heatmap" {
		return []float64{}
	}
	// default (heatmap queries don't compute percentiles)
	if len(lats) == 0 {
		return []float64{50, 95, 99, 99.9}
	}
	latencies := make([]float64, 0, len(lats))
	for _, l := range lats {
		latencies = append(latencies, l.(float64))
	}
	return latencies
}

func buildSpansQuery(spansQuery interface{}, display string, finalWindowOperation *client.FinalWindowOperation) client.SpansQuery {
	var sq client.SpansQuery
	if spansQuery == nil || len(spansQuery.([]interface{})) == 0 {
		return sq
	}
	s := spansQuery.([]interface{})[0].(map[string]interface{})
	sq.Query = s["query"].(string)
	sq.Operator = s["operator"].(string)

	operatorInputWindowMs := s["operator_input_window_ms"]
	if operatorInputWindowMs != 0 { // 0 here indicates that the value was not set in terraform file
		value := operatorInputWindowMs.(int)
		sq.OperatorInputWindowMs = &value
	}

	if sq.Operator == "latency" {
		sq.LatencyPercentiles = buildLatencyPercentiles(s["latency_percentiles"].([]interface{}), display)
	}
	if groupByKeys, ok := s["group_by_keys"].([]interface{}); ok && len(groupByKeys) > 0 {
		sq.GroupByKeys = buildSpansGroupByKeys(s["group_by_keys"].([]interface{}))
	}
	sq.FinalWindowOperation = finalWindowOperation

	return sq
}

func buildQueries(queriesIn []interface{}) ([]client.MetricQueryWithAttributes, error) {
	var newQueries []client.MetricQueryWithAttributes
	var queries []map[string]interface{}
	for _, queryIn := range queriesIn {
		queries = append(queries, queryIn.(map[string]interface{}))
	}

	hasSpanSingle := false
	for _, query := range queries {

		// When checking if this chart uses a query string, check deprecated TQL field as well
		queryString, ok := query["query_string"].(string)
		if !ok || queryString == "" {
			queryString, _ = query["tql"].(string)
		}

		if queryString != "" {
			newQuery := client.MetricQueryWithAttributes{
				Name:                 query["query_name"].(string),
				Type:                 "tql",
				Hidden:               query["hidden"].(bool),
				Display:              query["display"].(string),
				TQLQuery:             queryString,
				DependencyMapOptions: buildDependencyMapOptions(query["dependency_map_options"]),
			}
			newQueries = append(newQueries, newQuery)
			continue
		}

		// If this chart uses a span query
		spansQuery := query["spans"]
		if spansQuery != nil && len(spansQuery.([]interface{})) > 0 {
			hasSpanSingle = true
			err := validateSpansQuery(spansQuery)
			if err != nil {
				return nil, err
			}
			display := query["display"].(string)
			newQuery := client.MetricQueryWithAttributes{
				Name:       query["query_name"].(string),
				Type:       "spans_single",
				Hidden:     query["hidden"].(bool),
				Display:    display,
				SpansQuery: buildSpansQuery(spansQuery, display, buildFinalWindowOperation(query["final_window_operation"])),
			}
			newQueries = append(newQueries, newQuery)
			continue
		}

		// If this chart uses a metric query or composite query
		metric := query["metric"].(string)
		queryType := "single"
		if metric == "" {
			queryType = "composite"
		}
		newQuery := client.MetricQueryWithAttributes{
			Name:    query["query_name"].(string),
			Type:    queryType,
			Hidden:  query["hidden"].(bool),
			Display: query["display"].(string),
			Query: client.MetricQuery{
				TimeseriesOperator: query["timeseries_operator"].(string),
				Metric:             metric,
			},
		}

		timeseriesOperatorInputWindowMs := query["timeseries_operator_input_window_ms"]
		if timeseriesOperatorInputWindowMs != 0 { // 0 here indicates that the value was not set in terraform file
			value := timeseriesOperatorInputWindowMs.(int)
			newQuery.Query.TimeseriesOperatorInputWindowMs = &value
		}

		if newQuery.Type == "single" {
			newQuery.Query.FinalWindowOperation = buildFinalWindowOperation(query["final_window_operation"])
		} else {
			newQuery.CompositeQuery.FinalWindowOperation = buildFinalWindowOperation(query["final_window_operation"])
		}

		includes := query["include_filters"]
		if includes != nil {
			err := validateFilters(includes.([]interface{}), false)
			if err != nil {
				return nil, err
			}
		}

		excludes := query["exclude_filters"]
		if excludes != nil {
			err := validateFilters(excludes.([]interface{}), false)
			if err != nil {
				return nil, err
			}
		}

		allFilters := query["filters"]
		if allFilters != nil {
			err := validateFilters(allFilters.([]interface{}), true)
			if err != nil {
				return nil, err
			}
		}
		filters := buildLabelFilters(includes.([]interface{}), excludes.([]interface{}), allFilters.([]interface{}))
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
				client.GroupBy{
					Aggregation: g["aggregation_method"].(string),
					LabelKeys:   buildKeys(g["keys"].([]interface{})),
				}
		}
		newQueries = append(newQueries, newQuery)
	}

	if hasSpanSingle {
		for i := 0; i < len(newQueries); i++ {
			if newQueries[i].Type == "composite" {
				newQueries[i].Type = "spans_composite"
			}
		}
	}

	return newQueries, nil
}

func buildDependencyMapOptions(in interface{}) *client.DependencyMapOptions {
	if in == nil || len(in.([]interface{})) == 0 {
		return nil
	}

	options := in.([]interface{})[0].(map[string]interface{})
	scope := options["scope"].(string)
	mapType := options["map_type"].(string)

	return &client.DependencyMapOptions{
		Scope:   scope,
		MapType: mapType,
	}
}

func buildFinalWindowOperation(in interface{}) *client.FinalWindowOperation {
	if in == nil || len(in.([]interface{})) == 0 {
		return nil
	}

	finalWindowOperator := in.([]interface{})[0].(map[string]interface{})
	operator := finalWindowOperator["operator"].(string)
	inputWindowMs := finalWindowOperator["input_window_ms"].(int)
	return &client.FinalWindowOperation{
		Operator:      operator,
		InputWindowMs: inputWindowMs,
	}
}

func buildKeys(keysIn []interface{}) []string {
	var keys []string
	for _, k := range keysIn {
		keys = append(keys, k.(string))
	}
	return keys
}

func buildThresholds(d *schema.ResourceData) (client.Thresholds, error) {
	t := client.Thresholds{}

	critical := d.Get("expression.0.thresholds.0.critical")
	if critical != "" {
		c, err := strconv.ParseFloat(critical.(string), 64)
		if err != nil {
			return t, err
		}
		t.Critical = &c
	}

	warning := d.Get("expression.0.thresholds.0.warning")
	if warning != "" {
		w, err := strconv.ParseFloat(warning.(string), 64)
		if err != nil {
			return t, err
		}
		t.Warning = &w
	}

	return t, nil
}

func buildLabelFilters(includes []interface{}, excludes []interface{}, all []interface{}) []client.LabelFilter {
	var filters []client.LabelFilter

	if len(includes) > 0 {
		for _, includeFilter := range includes {
			key := includeFilter.(map[string]interface{})["key"]
			value := includeFilter.(map[string]interface{})["value"]

			filters = append(filters, client.LabelFilter{
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
			filters = append(filters, client.LabelFilter{
				Operand: "neq",
				Key:     key.(string),
				Value:   value.(string),
			})
		}
	}

	if len(all) > 0 {
		for _, allFilter := range all {
			key := allFilter.(map[string]interface{})["key"]
			value := allFilter.(map[string]interface{})["value"]
			operand := allFilter.(map[string]interface{})["operand"]
			filters = append(filters, client.LabelFilter{
				Operand: operand.(string),
				Key:     key.(string),
				Value:   value.(string),
			})
		}
	}
	return filters
}

func validateSpansQuery(spansQuery interface{}) error {
	s := spansQuery.([]interface{})[0].(map[string]interface{})
	query, hasQuery := s["query"]
	if !hasQuery {
		return fmt.Errorf("missing required field query om spans")
	}

	operator, hasOperator := s["operator"]
	if !hasOperator {
		return fmt.Errorf("missing required field operator on spans")
	}

	switch query.(type) {
	case string:
	default:
		return fmt.Errorf("value must be a string. got: %v", query)
	}

	switch operator.(type) {
	case string:
	default:
		return fmt.Errorf("value must be a string. got: %T", operator)
	}
	return nil
}

func validateFilters(filters []interface{}, hasOperand bool) error {
	for _, filter := range filters {
		key, ok := filter.(map[string]interface{})["key"]
		if !ok {
			return fmt.Errorf("'key' is a required field")
		}

		value, ok := filter.(map[string]interface{})["value"]
		if !ok {
			return fmt.Errorf("'value' is a required field")
		}

		if hasOperand {
			op, ok := filter.(map[string]interface{})["operand"]
			if !ok {
				return fmt.Errorf("'operand' is a required field")
			}
			switch op.(type) {
			case string:
			default:
				return fmt.Errorf("operand must be a string. got: %v", key)
			}

			if op.(string) == "eq" || op.(string) == "neq" {
				return fmt.Errorf("filters object does not support operand %s: use include_filters or exclude_filters instead", op)
			}
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

func setResourceDataFromUnifiedCondition(project string, c client.UnifiedCondition, d *schema.ResourceData, schemaType ConditionSchemaType) error {
	if err := d.Set("project_name", project); err != nil {
		return fmt.Errorf("unable to set project_name resource field: %v", err)
	}

	if err := d.Set("name", c.Attributes.Name); err != nil {
		return fmt.Errorf("unable to set name resource field: %v", err)
	}

	if err := d.Set("description", c.Attributes.Description); err != nil {
		return fmt.Errorf("unable to set description resource field: %v", err)
	}

	if err := d.Set("type", "metric_alert"); err != nil {
		return fmt.Errorf("unable to set type resource field: %v", err)
	}

	thresholdEntries := map[string]interface{}{}

	if c.Attributes.Expression.Thresholds.Critical != nil {
		thresholdEntries["critical"] = strconv.FormatFloat(*c.Attributes.Expression.Thresholds.Critical, 'f', -1, 64)
	}

	if c.Attributes.Expression.Thresholds.Warning != nil {
		thresholdEntries["warning"] = strconv.FormatFloat(*c.Attributes.Expression.Thresholds.Warning, 'f', -1, 64)
	}

	if err := d.Set("expression", []map[string]interface{}{
		{
			"is_multi":   c.Attributes.Expression.IsMulti,
			"is_no_data": c.Attributes.Expression.IsNoData,
			"operand":    c.Attributes.Expression.Operand,
			"thresholds": []interface{}{
				thresholdEntries,
			},
		},
	}); err != nil {
		return fmt.Errorf("unable to set expression resource field: %v", err)
	}

	if schemaType == MetricConditionSchema {
		if err := d.Set("metric_query", getQueriesFromMetricConditionData(c.Attributes.Queries)); err != nil {
			return fmt.Errorf("unable to set metric_query resource field: %v", err)
		}
	} else {
		queries, err := getQueriesFromUnifiedConditionResourceData(
			c.Attributes.Queries,
			c.ID,
		)
		if err != nil {
			return err
		}
		if err := d.Set("query", queries); err != nil {
			return fmt.Errorf("unable to set query resource field: %v", err)
		}
	}

	var alertingRules []interface{}
	for _, r := range c.Attributes.AlertingRules {
		includeFilters, excludeFilters, allFilters := getIncludeExcludeFilters(r.MatchOn.GroupBy)

		alertingRules = append(alertingRules, map[string]interface{}{
			"id":              r.MessageDestinationID,
			"update_interval": GetUpdateIntervalValue(r.UpdateInterval),
			"include_filters": includeFilters,
			"exclude_filters": excludeFilters,
			"filters":         allFilters,
		})
	}

	alertingRuleSet := schema.NewSet(
		schema.HashResource(&schema.Resource{
			Schema: getAlertingRuleSchema(),
		}),
		alertingRules,
	)
	if err := d.Set("alerting_rule", alertingRuleSet); err != nil {
		return fmt.Errorf("unable to set alerting_rule resource field: %v", err)
	}

	return nil
}

func getQueriesFromUnifiedConditionResourceData(
	queriesIn []client.MetricQueryWithAttributes,
	conditionID string,
) ([]interface{}, error) {
	var queries []interface{}
	for _, q := range queriesIn {
		if q.Type != "tql" {
			return nil, fmt.Errorf(
				"cannot convert query from condition %v\n\n"+
					"Query is of type '%v' but must be of type 'tql' for use with the resource\n"+
					"type lightstep_condition.\n"+
					"\n"+
					"Try using the lightstep_metrics_condition resource type for this condition\n"+
					"or convert all queries in the condition to query string format. ",
				conditionID,
				q.Type,
			)
		}

		qs := map[string]interface{}{
			"hidden":       q.Hidden,
			"display":      q.Display,
			"query_name":   q.Name,
			"query_string": q.TQLQuery,
		}
		queries = append(queries, qs)
	}
	return queries, nil
}

func getIncludeExcludeFilters(filters []client.LabelFilter) ([]interface{}, []interface{}, []interface{}) {
	var includeFilters []interface{}
	var excludeFilters []interface{}
	var allFilters []interface{}

	for _, f := range filters {

		// for backwards compatibility for v1.60.4 and below
		if f.Operand == "eq" {
			includeFilters = append(includeFilters, map[string]interface{}{
				"key":   f.Key,
				"value": f.Value,
			})
		} else if f.Operand == "neq" {
			excludeFilters = append(excludeFilters, map[string]interface{}{
				"key":   f.Key,
				"value": f.Value,
			})
		} else {
			allFilters = append(allFilters, map[string]interface{}{
				"key":     f.Key,
				"value":   f.Value,
				"operand": f.Operand,
			})
		}
	}
	return includeFilters, excludeFilters, allFilters
}

func getQueriesFromMetricConditionData(queriesIn []client.MetricQueryWithAttributes) []interface{} {
	var queries []interface{}
	for _, q := range queriesIn {
		includeFilters, excludeFilters, allFilters := getIncludeExcludeFilters(q.Query.Filters)

		var groupBy []interface{}
		if q.Query.GroupBy.Aggregation != "" || len(q.Query.GroupBy.LabelKeys) > 0 {
			groupBy = []interface{}{
				map[string]interface{}{
					"aggregation_method": q.Query.GroupBy.Aggregation,
					"keys":               q.Query.GroupBy.LabelKeys,
				},
			}
		}

		qs := map[string]interface{}{
			"metric":              q.Query.Metric,
			"hidden":              q.Hidden,
			"display":             q.Display,
			"query_name":          q.Name,
			"timeseries_operator": q.Query.TimeseriesOperator,
			"include_filters":     includeFilters,
			"exclude_filters":     excludeFilters,
			"filters":             allFilters,
			"group_by":            groupBy,
			"tql":                 q.TQLQuery, // deprecated
		}
		if q.Query.TimeseriesOperatorInputWindowMs != nil {
			qs["timeseries_operator_input_window_ms"] = *q.Query.TimeseriesOperatorInputWindowMs
		}
		if q.Query.FinalWindowOperation != nil {
			qs["final_window_operation"] = getFinalWindowOperationFromResourceData(q.Query.FinalWindowOperation)
		} else if q.CompositeQuery.FinalWindowOperation != nil {
			qs["final_window_operation"] = getFinalWindowOperationFromResourceData(q.CompositeQuery.FinalWindowOperation)
		} else if q.SpansQuery.FinalWindowOperation != nil {
			qs["final_window_operation"] = getFinalWindowOperationFromResourceData(q.SpansQuery.FinalWindowOperation)
		}

		if q.SpansQuery.Query != "" {
			sqi := map[string]interface{}{
				"query":                    q.SpansQuery.Query,
				"operator":                 q.SpansQuery.Operator,
				"operator_input_window_ms": q.SpansQuery.OperatorInputWindowMs,
			}
			if q.SpansQuery.OperatorInputWindowMs != nil {
				sqi["operator_input_window_ms"] = *q.SpansQuery.OperatorInputWindowMs
			}
			if len(q.SpansQuery.GroupByKeys) > 0 {
				sqi["group_by_keys"] = q.SpansQuery.GroupByKeys
			}
			if q.SpansQuery.Operator == "latency" {
				sqi["latency_percentiles"] = q.SpansQuery.LatencyPercentiles
			}

			qs["spans"] = []interface{}{
				sqi,
			}
		}
		queries = append(queries, qs)
	}
	return queries
}

func getFinalWindowOperationFromResourceData(finalWindowOperation *client.FinalWindowOperation) []interface{} {
	if finalWindowOperation == nil {
		return nil
	}

	return []interface{}{
		map[string]interface{}{
			"operator":        finalWindowOperation.Operator,
			"input_window_ms": finalWindowOperation.InputWindowMs,
		},
	}
}
