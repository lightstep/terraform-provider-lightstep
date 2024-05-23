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
// (2) The unified lightstep_alert
//
// The resources are largely the same with the primary differences being the
// query format and composite alert support.
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
				Type:        schema.TypeString,
				Required:    true,
				Description: "The name of the [project](https://docs.lightstep.com/docs/glossary#project) in which to create this alert.",
			},
			"name": {
				Type:        schema.TypeString,
				Required:    true,
				Description: "The title of the alert.",
			},
			"type": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"description": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: "Optional extended description for the alert (supports Markdown).",
			},
			"label": {
				Type:        schema.TypeSet,
				Optional:    true,
				Description: "Optional labels to attach to this alert. Labels can be key/value pairs or standalone values.",
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"key": {
							Type:     schema.TypeString,
							Optional: true,
						},
						"value": {
							Type:     schema.TypeString,
							Required: true,
						},
					},
				},
			},
			"custom_data": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: "Optional free-form string to include in alert notifications (max length 4096 bytes).",
			},
			"alerting_rule": {
				Type:        schema.TypeSet,
				Optional:    true,
				Description: "Optional configuration to receive alert notifications.",
				Elem: &schema.Resource{
					Schema: getAlertingRuleSchemaMap(),
				},
			},
		},
	}

	if conditionSchemaType == UnifiedConditionSchema {
		resource.Schema["expression"] = getUnifiedAlertExpressionSchema()
		resource.Schema["query"] = &schema.Schema{
			Type:        schema.TypeList,
			Optional:    true,
			Description: "Defines the query for a single alert. For a composite alert, use the composite_alert section instead.",
			Elem: &schema.Resource{
				Schema: getUnifiedQuerySchemaMap(),
			},
		}
		// Configuration for a composite alert, consists of two or more sub alerts
		resource.Schema["composite_alert"] = &schema.Schema{
			Type:        schema.TypeList,
			Optional:    true,
			MinItems:    1,
			MaxItems:    1,
			Description: "Defines the configuration for a [composite alert](https://docs.lightstep.com/docs/about-alerts#customize-alerts-with-alert-templates). Mutually exclusive with { query, expression } which define the configuration for a single alert.",
			Elem: &schema.Resource{
				Schema: map[string]*schema.Schema{
					"alert": {
						Type:        schema.TypeSet,
						Required:    true,
						MinItems:    1,
						MaxItems:    10,
						Description: "Defines one of the sub alerts within a composite alert.",
						Elem: &schema.Resource{
							Schema: getCompositeSubAlertSchemaMap(),
						},
					},
				},
			},
		}
	} else {
		// mark the whole resource as deprecated
		resource.DeprecationMessage = "This resource is deprecated. Please migrate to lightstep_alert."

		resource.Schema["expression"] = getMetricConditionExpressionSchema()
		resource.Schema["metric_query"] = &schema.Schema{
			Type:        schema.TypeList,
			Required:    true,
			Description: "Defines the alert query",
			Elem: &schema.Resource{
				Schema: getMetricQuerySchemaMap(),
			},
		}
	}
	return resource
}

func getAlertingRuleSchemaMap() map[string]*schema.Schema {
	return map[string]*schema.Schema{
		"update_interval": {
			Type:         schema.TypeString,
			Optional:     true,
			ValidateFunc: validation.StringInSlice(GetValidUpdateInterval(), false),
			Description: `An optional duration that represents the frequency at which to re-send an alert notification if an alert remains in a triggered state. 
By default, notifications will only be sent when the alert status changes.  
Values should be expressed as a duration (example: "2d").`,
		},
		"id": {
			Type:        schema.TypeString,
			Required:    true,
			Description: `The identifier of the destination to receive notifications for this alert.`,
		},
	}
}

func getSpansQuerySchema() *schema.Schema {
	sma := schema.Schema{
		Type:       schema.TypeList,
		MaxItems:   1,
		Deprecated: "This field and resource are deprecated. Please migrate to the lightstep_alerts resource type.",
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

func getMetricQuerySchemaMap() map[string]*schema.Schema {
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
			ValidateFunc: validation.StringInSlice([]string{"line", "area", "bar", "big_number", "heatmap", "dependency_map", "big_number_v2", "scatter_plot"}, false),
		},
		"query_name": {
			Type:         schema.TypeString,
			Required:     true,
			ValidateFunc: validation.StringLenBetween(1, 128),
		},
		"timeseries_operator": {
			Type:         schema.TypeString,
			Optional:     true,
			Computed:     true,
			ValidateFunc: validation.StringInSlice([]string{"rate", "delta", "last", "min", "max", "avg"}, false),
		},
		"timeseries_operator_input_window_ms": {
			Type:         schema.TypeInt,
			Description:  "Unit specified in milliseconds, but must be at least 30,000 and a round number of seconds (i.e. evenly divisible by 1,000).",
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
			Deprecated:  "Use lightstep_dashboard or lightstep_alert instead",
			Description: "Deprecated, use the query_string field in lightstep_dashboard or lightstep_alert instead",
			Type:        schema.TypeString,
			Optional:    true,
			Computed:    false,
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
					Description:  "Unit specified in milliseconds, but must be at least 30,000 and a round number of seconds (i.e. evenly divisible by 1,000).",
					Optional:     true,
					ValidateFunc: validation.All(validation.IntDivisibleBy(1_000), validation.IntAtLeast(30_000)),
				},
			},
		},
	}
}

func getThresholdSchemaMap() map[string]*schema.Schema {
	return map[string]*schema.Schema{
		"critical": {
			Type:        schema.TypeString,
			Optional:    true,
			Description: "Defines the threshold for the alert to transition to a Critical (more severe) status.",
		},
		"critical_duration_ms": {
			Type:        schema.TypeInt,
			Optional:    true,
			Description: "Critical threshold must be breached for this duration before the status changes.",
		},
		"warning": {
			Type:        schema.TypeString,
			Optional:    true,
			Description: "Defines the threshold for the alert to transition to a Warning (less severe) status.",
		},
		"warning_duration_ms": {
			Type:        schema.TypeInt,
			Optional:    true,
			Description: "Critical threshold must be breached for this duration before the status changes.",
		},
	}
}

// expression is optional because it cannot be included in composite alerts but is otherwise required
func getUnifiedAlertExpressionSchema() *schema.Schema {
	resource := getCompositeSubAlertExpressionResource()
	resource.Schema["is_multi"] = getIsMultiSchema()
	return &schema.Schema{
		Type:        schema.TypeList,
		Optional:    true,
		Description: "Describes the conditions that trigger a single alert. For a composite alert, use the composite_alert section instead.",
		MaxItems:    1,
		Elem:        resource,
	}
}

// expression is required in legacy metric conditions
func getMetricConditionExpressionSchema() *schema.Schema {
	resource := getCompositeSubAlertExpressionResource()
	resource.Schema["is_multi"] = getIsMultiSchema()
	return &schema.Schema{
		Type:        schema.TypeList,
		Required:    true,
		MaxItems:    1,
		MinItems:    1,
		Description: "Describes the conditions that trigger the alert.",
		Elem:        resource,
	}
}

func getCompositeSubAlertSchemaMap() map[string]*schema.Schema {
	return map[string]*schema.Schema{
		"name": {
			Type:        schema.TypeString,
			Required:    true,
			Description: `The identifier for this sub alert. Must be a single uppercase letter (examples: "A", "B", "C")`,
		},
		"title": {
			Type:        schema.TypeString,
			Optional:    true,
			Default:     "",
			Description: "Optional free-form title for this sub alert.",
		},
		"expression": getCompositeSubAlertExpressionSchema(),
		"query": {
			Type:     schema.TypeList,
			Required: true,
			MaxItems: 1,
			MinItems: 1,
			Elem: &schema.Resource{
				Schema: getUnifiedQuerySchemaMap(),
			},
		},
	}
}

func getCompositeSubAlertExpressionSchema() *schema.Schema {
	return &schema.Schema{
		Type:     schema.TypeList,
		Required: true,
		MaxItems: 1,
		MinItems: 1,
		Elem:     getCompositeSubAlertExpressionResource(),
	}
}

// composite sub alerts do not offer multi alerting
func getCompositeSubAlertExpressionResource() *schema.Resource {
	return &schema.Resource{
		Schema: map[string]*schema.Schema{
			// The expression must either include `is_no_data = true` OR an operand and a threshold.
			// However that logic can't be expressed statically using the Required attribute so we
			// just mark all these fields as optional and let the server handle the detailed validation.
			"is_no_data": {
				Type:        schema.TypeBool,
				Optional:    true,
				Default:     false,
				Description: "If true, a notification is sent when the alert query returns no data. If false, notifications aren't sent in this scenario.",
			},
			"no_data_duration_ms": {
				Type:        schema.TypeInt,
				Optional:    true,
				Description: "No data must be seen for this duration before the status changes.",
			},
			"operand": {
				Type:         schema.TypeString,
				Optional:     true,
				ValidateFunc: validation.StringInSlice([]string{"", "above", "below"}, false),
				Description:  "Required when at least one threshold (Critical, Warning) is defined. Indicates whether the alert triggers when the value is above the threshold or below the threshold.",
			},
			"thresholds": {
				Type:     schema.TypeList,
				Optional: true,
				DiffSuppressFunc: func(k, old, new string, d *schema.ResourceData) bool {
					if strings.HasSuffix(k, "thresholds.#") {
						// When parsing the resource that we get from the Lightstep API, we don't include the thresholds
						// block unless at least one of the thresholds is set. If a user includes an empty thresholds
						// block in their config instead of omitting it entirely, terraform would normally treat that as a diff.
						// However, for convenience, we allow users to specify an empty thresholds block and
						// we treat it as semantically equivalent to omitting the thresholds block entirely.
						oldIntf, newIntf := d.GetChange(strings.TrimRight(k, ".#"))
						newList := newIntf.([]interface{})
						if len(oldIntf.([]interface{})) == 0 && len(newList) == 1 && newList[0] == nil {
							return true
						}
					}
					return false // evaluate diffs normally
				},
				MaxItems:    1,
				MinItems:    0,
				Description: "Optional values defining the thresholds at which this alert transitions into Critical or Warning states. If a particular threshold is not specified, the alert never transitions into that state.",
				Elem: &schema.Resource{
					Schema: getThresholdSchemaMap(),
				},
			},
		},
	}
}

func getIsMultiSchema() *schema.Schema {
	return &schema.Schema{
		Type:        schema.TypeBool,
		Optional:    true,
		Default:     false,
		Description: "When false, send a single notification whenever any number of group_by values exceeds the alert threshold. When true, send individual notifications for each distinct group_by value that exceeds the threshold.",
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

	// Support for deprecated legacy queries: if we created a new legacy query and the creation
	// succeeded, return the ResourceData "as-is" from what was passed in. This avoids meaningless
	// diffs in the plan.
	projectName := d.Get("project_name").(string)
	legacy, err := metricConditionHasEquivalentLegacyQueries(ctx, c, projectName, attributes, &created.Attributes)
	if err != nil {
		return diag.FromErr(fmt.Errorf("failed to compare legacy queries: %v", err))
	}
	if legacy {
		created.Attributes.Queries = attributes.Queries
		if err := setResourceDataFromUnifiedCondition(projectName, created, d, p.conditionSchemaType); err != nil {
			return diag.FromErr(fmt.Errorf("failed to set condition from API response to terraform state: %v", err))
		}
		return nil
	}

	return p.resourceUnifiedConditionRead(ctx, d, m)
}

func (p *resourceUnifiedConditionImp) resourceUnifiedConditionRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	var diags diag.Diagnostics

	c := m.(*client.Client)
	prevAttrs, err := getUnifiedConditionAttributesFromResource(d, p.conditionSchemaType)
	if err != nil {
		return diag.FromErr(fmt.Errorf("failed to translate resource attributes: %v", err))
	}

	projectName := d.Get("project_name").(string)
	cond, err := c.GetUnifiedCondition(ctx, projectName, d.Id())
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

	legacy, err := metricConditionHasEquivalentLegacyQueries(ctx, c, projectName, prevAttrs, &cond.Attributes)
	if err != nil {
		return diag.FromErr(fmt.Errorf("failed to compare legacy queries: %v", err))
	}
	if legacy {
		cond.Attributes.Queries = prevAttrs.Queries
	}

	err = setResourceDataFromUnifiedCondition(projectName, *cond, d, p.conditionSchemaType)
	if err != nil {
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
	var (
		expression *client.Expression
		err        error
	)
	expressionList := d.Get("expression").([]interface{})
	if len(expressionList) > 0 {
		expression, err = buildExpression(expressionList[0].(map[string]interface{}))
		if err != nil {
			return nil, err
		}
	}

	labelSet := d.Get("label").(*schema.Set)
	labels, err := buildLabels(labelSet.List())
	if err != nil {
		return nil, err
	}

	schemaQuery := d.Get("metric_query")
	if schemaType == UnifiedConditionSchema {
		schemaQuery = d.Get("query")
	}
	queries, err := buildQueries(schemaQuery.([]interface{}))
	if err != nil {
		return nil, err
	}

	alertingRules, err := buildAlertingRules(d.Get("alerting_rule").(*schema.Set))
	if err != nil {
		return nil, err
	}

	compositeAlert, err := buildCompositeAlert(d)
	if err != nil {
		return nil, err
	}

	return &client.UnifiedConditionAttributes{
		Type:           "metrics",
		Name:           d.Get("name").(string),
		Description:    d.Get("description").(string),
		Expression:     expression,
		Labels:         labels,
		CustomData:     d.Get("custom_data").(string),
		AlertingRules:  alertingRules,
		Queries:        queries,
		CompositeAlert: compositeAlert,
	}, nil
}

func buildExpression(singleExpression map[string]interface{}) (*client.Expression, error) {
	subalertExpression, err := buildSubAlertExpression(singleExpression)
	if err != nil {
		return nil, err
	}

	return &client.Expression{
		IsMulti:            singleExpression["is_multi"].(bool),
		SubAlertExpression: *subalertExpression,
	}, nil
}

func buildSubAlertExpression(singleExpression map[string]interface{}) (*client.SubAlertExpression, error) {
	thresholds, err := buildThresholds(singleExpression)
	if err != nil {
		return nil, err
	}

	e := &client.SubAlertExpression{
		IsNoData:   singleExpression["is_no_data"].(bool),
		Operand:    singleExpression["operand"].(string),
		Thresholds: thresholds,
	}

	noDataDuration := singleExpression["no_data_duration_ms"]
	if noDataDuration != nil && noDataDuration != "" && noDataDuration != 0 {
		d, ok := noDataDuration.(int)
		if !ok {
			return e, err
		}
		e.NoDataDurationMs = &d
	}
	return e, nil
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

		updateIntervalMilli, ok := validUpdateInterval[rule["update_interval"].(string)]
		if ok {
			newRule.UpdateInterval = updateIntervalMilli
		}

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
				QueryString:          queryString,
				DependencyMapOptions: buildDependencyMapOptions(query["dependency_map_options"]),
			}

			// Check for the optional JSON block of display options
			if opts, ok := query["display_type_options"].(*schema.Set); ok {
				list := opts.List()
				count := len(list)
				if count > 1 {
					return nil, fmt.Errorf("display_type_options must be defined only once")
				} else if count == 1 {
					m, ok := list[0].(map[string]interface{})
					if !ok {
						return nil, fmt.Errorf("unexpected format for display_type_options")
					}
					// The API treats display_type_options as an opaque blob, so now that
					// we can pass what we have along directly
					newQuery.DisplayTypeOptions = m
				}
			}

			// "hidden_queries" is only applicable to "tql"/ "query_string" queries.
			// Due to Terraform's issues with TypeMap of TypeBool, we're forced to use strings
			if hiddenQueries, ok := query["hidden_queries"].(map[string]interface{}); ok && len(hiddenQueries) > 0 {
				hq := make(map[string]bool, len(hiddenQueries))
				for k, v := range hiddenQueries {
					s, ok := v.(string)
					if ok && s == "true" {
						hq[k] = true
					} else {
						hq[k] = false
					}
				}
				newQuery.HiddenQueries = hq
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

func buildThresholds(singleExpression map[string]interface{}) (client.Thresholds, error) {
	t := client.Thresholds{}

	elem, ok := singleExpression["thresholds"]
	if !ok {
		return t, nil
	}
	elemList := elem.([]interface{})
	if len(elemList) == 0 || elemList[0] == nil {
		return t, nil
	}
	thresholdsObj := elemList[0].(map[string]interface{})

	critical := thresholdsObj["critical"]
	if critical != "" {
		c, err := strconv.ParseFloat(critical.(string), 64)
		if err != nil {
			return t, err
		}
		t.Critical = &c
	}

	warning := thresholdsObj["warning"]
	if warning != "" {
		w, err := strconv.ParseFloat(warning.(string), 64)
		if err != nil {
			return t, err
		}
		t.Warning = &w
	}

	criticalDuration := thresholdsObj["critical_duration_ms"]
	if criticalDuration != nil && criticalDuration != "" && criticalDuration != 0 {
		d, ok := criticalDuration.(int)
		if !ok {
			return t, fmt.Errorf("unexpected format for critical_duration_ms")
		}
		t.CriticalDurationMs = &d
	}

	warningDuration := thresholdsObj["warning_duration_ms"]
	if warningDuration != nil && warningDuration != "" && warningDuration != 0 {
		d, ok := warningDuration.(int)
		if !ok {
			return t, fmt.Errorf("unexpected format for warning_duration_ms")
		}
		t.WarningDurationMs = &d
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

func buildCompositeAlert(d *schema.ResourceData) (*client.CompositeAlert, error) {
	compositeAlertInUntyped := d.Get("composite_alert")
	if compositeAlertInUntyped == nil {
		return nil, nil
	}

	compositeAlertIn := compositeAlertInUntyped.([]interface{})
	if len(compositeAlertIn) == 0 {
		return nil, nil
	}

	subAlertsInUntyped, ok := compositeAlertIn[0].(map[string]interface{})
	if !ok {
		return nil, fmt.Errorf("could not parse composite_alert")
	}

	subAlertsIn := subAlertsInUntyped["alert"].(*schema.Set).List()
	subAlerts := make([]client.CompositeSubAlert, 0, len(subAlertsIn))

	for _, subAlertInUntyped := range subAlertsIn {

		subAlertIn, ok := subAlertInUntyped.(map[string]interface{})
		if !ok {
			return nil, fmt.Errorf("could not parse alert")
		}

		subAlertExpression, err := buildSubAlertExpression(subAlertIn["expression"].([]interface{})[0].(map[string]interface{}))
		if err != nil {
			return nil, err
		}

		subAlertQueries, err := buildQueries(subAlertIn["query"].([]interface{}))
		if err != nil {
			return nil, err
		}
		if len(subAlertQueries) != 1 {
			return nil, fmt.Errorf("each sub alert requires exactly one query")
		}

		subAlert := client.CompositeSubAlert{
			Name:       subAlertIn["name"].(string),
			Title:      subAlertIn["title"].(string),
			Expression: *subAlertExpression,
			Queries:    subAlertQueries,
		}
		subAlerts = append(subAlerts, subAlert)
	}

	return &client.CompositeAlert{
		Alerts: subAlerts,
	}, nil
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

	labels := extractLabels(c.Attributes.Labels)
	if err := d.Set("label", labels); err != nil {
		return fmt.Errorf("unable to set labels resource field: %v", err)
	}

	if err := d.Set("custom_data", c.Attributes.CustomData); err != nil {
		return fmt.Errorf("unable to set type resource field: %v", err)
	}

	if err := d.Set("type", "metric_alert"); err != nil {
		return fmt.Errorf("unable to set type resource field: %v", err)
	}

	if c.Attributes.Expression != nil {
		expressionMap := map[string]interface{}{
			"is_multi":   c.Attributes.Expression.IsMulti,
			"is_no_data": c.Attributes.Expression.IsNoData,
			"operand":    c.Attributes.Expression.Operand,
			"thresholds": buildUntypedThresholds(c.Attributes.Expression.Thresholds),
		}
		if c.Attributes.Expression.NoDataDurationMs != nil {
			expressionMap["no_data_duration_ms"] = c.Attributes.Expression.NoDataDurationMs
		}
		expressionSlice := []map[string]interface{}{expressionMap}
		if err := d.Set("expression", expressionSlice); err != nil {
			return fmt.Errorf("unable to set expression resource field: %v", err)
		}
	}

	if schemaType == MetricConditionSchema {
		if err := d.Set("metric_query", getQueriesFromMetricConditionData(c.Attributes.Queries)); err != nil {
			return fmt.Errorf("unable to set metric_query resource field: %v", err)
		}
	} else {
		queries, err := getQueriesFromUnifiedConditionResourceData(
			c.Attributes.Queries,
			c.ID,
			"",
		)
		if err != nil {
			return err
		}
		if err := d.Set("query", queries); err != nil {
			return fmt.Errorf("unable to set query resource field: %v", err)
		}

		if c.Attributes.CompositeAlert != nil {
			compositeAlert, err := getCompositeAlertFromUnifiedConditionResourceData(c.Attributes.CompositeAlert)
			if err != nil {
				return err
			}

			if err = d.Set("composite_alert", compositeAlert); err != nil {
				return fmt.Errorf("unable to set composite_alert field: %s", err)
			}
		}
	}

	var alertingRules []interface{}
	for _, r := range c.Attributes.AlertingRules {
		alertingRules = append(alertingRules, map[string]interface{}{
			"id":              r.MessageDestinationID,
			"update_interval": GetUpdateIntervalValue(r.UpdateInterval),
		})
	}

	alertingRuleSet := schema.NewSet(
		schema.HashResource(&schema.Resource{
			Schema: getAlertingRuleSchemaMap(),
		}),
		alertingRules,
	)
	if err := d.Set("alerting_rule", alertingRuleSet); err != nil {
		return fmt.Errorf("unable to set alerting_rule resource field: %v", err)
	}

	return nil
}

func buildUntypedThresholds(thresholds client.Thresholds) []map[string]interface{} {
	if thresholds.Warning == nil && thresholds.Critical == nil {
		return nil
	}

	outputMap := map[string]interface{}{}
	if thresholds.Critical != nil {
		outputMap["critical"] = strconv.FormatFloat(*thresholds.Critical, 'f', -1, 64)
	}

	if thresholds.Warning != nil {
		outputMap["warning"] = strconv.FormatFloat(*thresholds.Warning, 'f', -1, 64)
	}

	if thresholds.CriticalDurationMs != nil {
		outputMap["critical_duration_ms"] = thresholds.CriticalDurationMs
	}
	if thresholds.WarningDurationMs != nil {
		outputMap["warning_duration_ms"] = thresholds.WarningDurationMs
	}
	return []map[string]interface{}{
		outputMap,
	}
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
			"tql":                 q.QueryString,
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
