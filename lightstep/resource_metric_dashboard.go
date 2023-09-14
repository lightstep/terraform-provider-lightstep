package lightstep

import (
	"context"
	"fmt"
	"net/http"
	"strings"

	"github.com/lightstep/terraform-provider-lightstep/client"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
)

type ChartSchemaType int

const (
	MetricChartSchema ChartSchemaType = iota
	UnifiedChartSchema
)

// resourceUnifiedDashboard creates a resource for either:
//
// (1) The legacy lightstep_metric_dashboard
// (2) The unified lightstep_dashboard
//
// The resources are largely the same with the primary difference being the
// query format.
func resourceUnifiedDashboard(chartSchemaType ChartSchemaType) *schema.Resource {
	p := resourceUnifiedDashboardImp{chartSchemaType: chartSchemaType}

	return &schema.Resource{
		CreateContext: p.resourceUnifiedDashboardCreate,
		ReadContext:   p.resourceUnifiedDashboardRead,
		UpdateContext: p.resourceUnifiedDashboardUpdate,
		DeleteContext: p.resourceUnifiedDashboardDelete,
		Importer: &schema.ResourceImporter{
			StateContext: p.resourceUnifiedDashboardImport,
		},
		Schema: map[string]*schema.Schema{
			"project_name": {
				Type:     schema.TypeString,
				Required: true,
			},
			"dashboard_name": {
				Type:     schema.TypeString,
				Required: true,
			},
			"dashboard_description": {
				Type:     schema.TypeString,
				Optional: true,
			},
			"type": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"chart": {
				Type:     schema.TypeSet,
				Optional: true,
				Elem: &schema.Resource{
					Schema: getChartSchema(chartSchemaType),
				},
			},
			"group": {
				Type:     schema.TypeSet,
				Optional: true,
				Elem: &schema.Resource{
					Schema: getGroupSchema(chartSchemaType),
				},
			},
			"label": {
				Type:        schema.TypeSet,
				Optional:    true,
				Description: "Labels can be key/value pairs or standalone values.",
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
			"template_variable": {
				Type:     schema.TypeSet,
				Optional: true,
				Elem: &schema.Resource{
					Schema: getTemplateVariableSchema(),
				},
				Description: "Variable to be used in dashboard queries for dynamically filtering telemetry data",
			},
		},
	}
}

func getGroupSchema(chartSchemaType ChartSchemaType) map[string]*schema.Schema {
	return map[string]*schema.Schema{
		"id": {
			Type:     schema.TypeString,
			Computed: true,
		},
		"rank": {
			Type:         schema.TypeInt,
			ValidateFunc: validation.IntAtLeast(0),
			Required:     true,
		},
		"title": {
			Type:     schema.TypeString,
			Optional: true,
		},
		"visibility_type": {
			Type:         schema.TypeString,
			ValidateFunc: validation.StringInSlice([]string{"implicit", "explicit"}, false),
			Required:     true,
		},
		"chart": {
			Type:     schema.TypeSet,
			Optional: true,
			Computed: true, // the charts can be mutated individually; chart mutations should not trigger group updates
			Elem: &schema.Resource{
				Schema: getChartSchema(chartSchemaType),
			},
		},
		"text_panel": {
			Type:     schema.TypeList,
			Optional: true,
			Computed: true, // the panels can be mutated individually; chart mutations should not trigger group updates
			Elem: &schema.Resource{
				Schema: getTextPanelSchema(),
			},
		},
	}
}

func getTextPanelSchema() map[string]*schema.Schema {
	return mergeSchemas(
		getPanelSchema(false),
		map[string]*schema.Schema{
			"text": {
				Type:     schema.TypeString,
				Required: true,
			},
		},
	)
}

// getPanelSchema returns the common metadata on any dashboard panel (timeseries charts or text panels)
//
// Timeseries charts requires "name" to be a required, while text panels don't.
func getPanelSchema(isNameRequired bool) map[string]*schema.Schema {
	nameSchema := func() *schema.Schema {
		if isNameRequired {
			return &schema.Schema{
				Type:     schema.TypeString,
				Required: true,
			}
		}

		return &schema.Schema{
			Type:     schema.TypeString,
			Optional: true,
			Default:  "",
		}
	}

	return map[string]*schema.Schema{
		// Alias for what we refer to as title elsewhere
		"name": nameSchema(),
		"description": {
			Type:     schema.TypeString,
			Optional: true,
			Default:  "",
		},
		"x_pos": {
			Type:         schema.TypeInt,
			ValidateFunc: validation.IntAtLeast(0),
			Default:      0,
			Optional:     true,
		},
		"y_pos": {
			Type:         schema.TypeInt,
			ValidateFunc: validation.IntAtLeast(0),
			Default:      0,
			Optional:     true,
		},
		"width": {
			Type:         schema.TypeInt,
			ValidateFunc: validation.IntAtLeast(0),
			Default:      0,
			Optional:     true,
		},
		"height": {
			Type:         schema.TypeInt,
			ValidateFunc: validation.IntAtLeast(0),
			Default:      0,
			Optional:     true,
		},
		"id": {
			Type:     schema.TypeString,
			Computed: true,
		},
	}
}

func getChartSchema(chartSchemaType ChartSchemaType) map[string]*schema.Schema {
	var querySchema map[string]*schema.Schema
	if chartSchemaType == UnifiedChartSchema {
		querySchema = getUnifiedQuerySchemaMap()
		querySchema["dependency_map_options"] = &schema.Schema{
			Type:     schema.TypeList,
			Optional: true,
			MaxItems: 1,
			Elem: &schema.Resource{
				Schema: map[string]*schema.Schema{
					"scope": {
						Type:         schema.TypeString,
						Optional:     true,
						ValidateFunc: validation.StringInSlice([]string{"all", "upstream", "downstream", "immediate"}, false),
					},
					"map_type": {
						Type:         schema.TypeString,
						Optional:     true,
						ValidateFunc: validation.StringInSlice([]string{"service", "operation"}, false),
					},
				},
			},
		}
	} else {
		querySchema = getMetricQuerySchemaMap()
	}

	return mergeSchemas(
		getPanelSchema(true),
		map[string]*schema.Schema{
			"type": {
				Type:         schema.TypeString,
				Required:     true,
				ValidateFunc: validation.StringInSlice([]string{"timeseries"}, true),
			},
			"rank": {
				Type:         schema.TypeInt,
				ValidateFunc: validation.IntAtLeast(0),
				Required:     true,
			},
			"y_axis": {
				Type:       schema.TypeList,
				MaxItems:   1,
				Deprecated: "The y_axis field is no longer used",
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"min": {
							Type:     schema.TypeFloat,
							Required: true,
						},
						"max": {
							Type:     schema.TypeFloat,
							Required: true,
						},
					},
				},
				Optional: true,
			},
			"query": {
				Type:     schema.TypeList,
				Required: true,
				Elem: &schema.Resource{
					Schema: querySchema,
				},
			},
			"subtitle": {
				Type:         schema.TypeString,
				Description:  "Subtitle to show beneath big number, unused in other chart types",
				Optional:     true,
				ValidateFunc: validation.StringLenBetween(0, 37),
			},
		},
	)
}

func getTemplateVariableSchema() map[string]*schema.Schema {
	return map[string]*schema.Schema{
		"name": {
			Type:        schema.TypeString,
			Required:    true,
			Description: "Unique (per dashboard) name for template variable, beginning with a letter or underscore and only containing letters, numbers, and underscores",
		},
		"suggestion_attribute_key": {
			Type:        schema.TypeString,
			Required:    true,
			Description: "Attribute key used as source for suggested template variable values appearing in Lightstep UI",
		},
		"default_values": {
			Type:     schema.TypeList,
			Required: true,
			Elem: &schema.Schema{
				Type: schema.TypeString,
			},
			Description: "One or more values to set the template variable to by default (if none are provided, defaults to all possible values)",
		},
	}
}

type resourceUnifiedDashboardImp struct {
	chartSchemaType ChartSchemaType
}

func (p *resourceUnifiedDashboardImp) resourceUnifiedDashboardCreate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	c := m.(*client.Client)
	attrs, hasLegacyChartsIn, err := getUnifiedDashboardAttributesFromResource(d)
	if err != nil {
		return diag.FromErr(fmt.Errorf("failed to get dashboard attributes: %v", err))
	}

	dashboard := client.UnifiedDashboard{
		Type:       "dashboard",
		Attributes: *attrs,
	}

	created, err := c.CreateUnifiedDashboard(ctx, d.Get("project_name").(string), dashboard)
	if err != nil {
		return diag.FromErr(fmt.Errorf("failed to create dashboard: %v", err))
	}

	d.SetId(created.ID)

	// Support for deprecated legacy queries: if we created a new legacy query and the creation
	// succeeded, return the ResourceData "as-is" from what was passed in. This avoids meaningless
	// diffs in the plan.
	projectName := d.Get("project_name").(string)
	legacy, err := dashboardHasEquivalentLegacyQueries(ctx, c, projectName, attrs.Charts, created.Attributes.Charts)
	if err != nil {
		return diag.FromErr(fmt.Errorf("failed to compare legacy queries: %v", err))
	}
	if legacy {
		// Only copy the query attributes
		for _, chart := range attrs.Charts {
			for j, d := range dashboard.Attributes.Charts {
				if d.Rank == chart.Rank {
					dashboard.Attributes.Charts[j].MetricQueries = chart.MetricQueries
				}
			}
		}
		if err := p.setResourceDataFromUnifiedDashboard(projectName, dashboard, d, hasLegacyChartsIn); err != nil {
			return diag.FromErr(fmt.Errorf("failed to set dashboard from API response to terraform state: %v", err))
		}
		return nil
	}

	return p.resourceUnifiedDashboardRead(ctx, d, m)
}

func (p *resourceUnifiedDashboardImp) resourceUnifiedDashboardRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	c := m.(*client.Client)

	prevAttrs, hasLegacyChartsIn, err := getUnifiedDashboardAttributesFromResource(d)
	if err != nil {
		return diag.FromErr(fmt.Errorf("failed to translate resource attributes: %v", err))
	}

	dashboard, err := c.GetUnifiedDashboard(ctx, d.Get("project_name").(string), d.Id())
	if err != nil {
		apiErr, ok := err.(client.APIResponseCarrier)
		if !ok {
			return diag.FromErr(fmt.Errorf("failed to get dashboard: %v", err))
		}

		if apiErr.GetStatusCode() == http.StatusNotFound {
			d.SetId("")
			return diags
		}

		return diag.FromErr(fmt.Errorf("failed to get dashboard: %v", apiErr))
	}

	// Support for deprecated legacy queries: if we created a new legacy query and the creation
	// succeeded, return the ResourceData "as-is" from what was passed in. This avoids false
	// diffs in the plan.
	projectName := d.Get("project_name").(string)
	legacyCharts, err := dashboardHasEquivalentLegacyQueries(ctx, c, projectName, prevAttrs.Charts, dashboard.Attributes.Charts)
	if err != nil {
		return diag.FromErr(fmt.Errorf("failed to compare legacy queries: %v", err))
	}
	if legacyCharts {
		// Only copy the query attributes
		for _, chart := range prevAttrs.Charts {
			for j, d := range dashboard.Attributes.Charts {
				if d.Rank == chart.Rank {
					dashboard.Attributes.Charts[j].MetricQueries = chart.MetricQueries
				}
			}
		}
	}

	for i, group := range prevAttrs.Groups {
		for j, g := range dashboard.Attributes.Groups {
			if g.Rank == group.Rank {
				previousGroup := prevAttrs.Groups[i]
				updatedGroup := dashboard.Attributes.Groups[j]
				legacyGroupedCharts, err := dashboardHasEquivalentLegacyQueries(
					ctx, c, projectName,
					previousGroup.Charts, updatedGroup.Charts)
				if err != nil {
					return diag.FromErr(fmt.Errorf("failed to compare legacy queries in groups: %v", err))
				}
				if legacyGroupedCharts {
					// Only copy the query attributes
					for _, chart := range previousGroup.Charts {
						for j, d := range updatedGroup.Charts {
							if d.Rank == chart.Rank {
								updatedGroup.Charts[j].MetricQueries = chart.MetricQueries
							}
						}
					}
				}
			}
		}
	}

	if err := p.setResourceDataFromUnifiedDashboard(d.Get("project_name").(string), *dashboard, d, hasLegacyChartsIn); err != nil {
		return diag.FromErr(fmt.Errorf("failed to set dashboard from API response to terraform state: %v", err))
	}
	return diags
}

func getUnifiedDashboardAttributesFromResource(d *schema.ResourceData) (*client.UnifiedDashboardAttributes, bool, error) {
	chartSet := d.Get("chart").(*schema.Set)
	groupSet := d.Get("group").(*schema.Set)
	groups, hasLegacyChartsIn, err := buildGroups(groupSet.List(), chartSet.List())
	if err != nil {
		return nil, hasLegacyChartsIn, err
	}

	labelSet := d.Get("label").(*schema.Set)
	labels, err := buildLabels(labelSet.List())
	if err != nil {
		return nil, hasLegacyChartsIn, err
	}

	templateVariableSet := d.Get("template_variable").(*schema.Set)
	templateVariables := buildTemplateVariables(templateVariableSet.List())

	attributes := &client.UnifiedDashboardAttributes{
		Name:              d.Get("dashboard_name").(string),
		Description:       d.Get("dashboard_description").(string),
		Groups:            groups,
		Labels:            labels,
		TemplateVariables: templateVariables,
	}

	return attributes, hasLegacyChartsIn, nil
}

func buildGroups(groupsIn []interface{}, legacyChartsIn []interface{}) ([]client.UnifiedGroup, bool, error) {
	var (
		newGroups         []client.UnifiedGroup
		hasLegacyChartsIn bool
	)

	if len(legacyChartsIn) != 0 {
		hasLegacyChartsIn = true
		c, err := buildCharts(legacyChartsIn)
		if err != nil {
			return nil, hasLegacyChartsIn, err
		}
		newGroups = append(newGroups, client.UnifiedGroup{
			Rank:           0,
			Title:          "",
			VisibilityType: "implicit",
			Charts:         c,
		})
	}

	for i := range groupsIn {
		group := groupsIn[i].(map[string]interface{})

		chartPanels, err := buildCharts(group["chart"].(*schema.Set).List())
		if err != nil {
			return nil, hasLegacyChartsIn, err
		}
		textPanels, err := buildTextPanels(group["text_panel"].([]interface{}))
		if err != nil {
			return nil, hasLegacyChartsIn, err
		}

		g := client.UnifiedGroup{
			ID:             group["id"].(string),
			Rank:           group["rank"].(int),
			Title:          group["title"].(string),
			VisibilityType: group["visibility_type"].(string),
			Charts:         append(chartPanels, textPanels...),
		}
		newGroups = append(newGroups, g)
	}
	return newGroups, hasLegacyChartsIn, nil
}

func buildPosition(panel map[string]interface{}) client.UnifiedPosition {
	return client.UnifiedPosition{
		XPos:   panel["x_pos"].(int),
		YPos:   panel["y_pos"].(int),
		Width:  panel["width"].(int),
		Height: panel["height"].(int),
	}
}

func buildTextPanels(textPanelsIn []interface{}) ([]client.UnifiedChart, error) {
	newCharts := []client.UnifiedChart{}
	for i := range textPanelsIn {
		textPanel := textPanelsIn[i].(map[string]interface{})
		c := client.UnifiedChart{
			ChartType: "text",
			ID:        textPanel["id"].(string),
			Title:     textPanel["name"].(string),
			Rank:      0, // Always 0, ignored for text panels
			Position:  buildPosition(textPanel),
			Text:      textPanel["text"].(string),
		}
		newCharts = append(newCharts, c)
	}
	return newCharts, nil
}

func buildCharts(chartsIn []interface{}) ([]client.UnifiedChart, error) {
	var (
		charts    []map[string]interface{}
		newCharts []client.UnifiedChart
	)

	for _, chart := range chartsIn {
		charts = append(charts, chart.(map[string]interface{}))
	}

	for _, chart := range charts {
		c := client.UnifiedChart{
			Title:       chart["name"].(string),
			Description: chart["description"].(string),
			Rank:        chart["rank"].(int),
			Position:    buildPosition(chart),
			ID:          chart["id"].(string),
			ChartType:   chart["type"].(string),
		}

		queries, err := buildQueries(chart["query"].([]interface{}))
		if err != nil {
			return nil, err
		}
		c.MetricQueries = queries

		yaxis, err := buildYAxis(chart["y_axis"].([]interface{}))
		if err != nil {
			return nil, err
		}
		if yaxis != nil {
			c.YAxis = yaxis
		}

		if subtitle, hasSubtitle := chart["subtitle"]; hasSubtitle {
			subtitleStr := subtitle.(string)
			c.Subtitle = &subtitleStr
		}

		newCharts = append(newCharts, c)
	}
	return newCharts, nil
}

func buildYAxis(yAxisIn []interface{}) (*client.YAxis, error) {
	if len(yAxisIn) < 1 {
		return nil, nil
	}
	y := yAxisIn[0].(map[string]interface{})

	max, ok := y["max"].(float64)
	if !ok {
		return nil, fmt.Errorf("missing required attribute 'max' for y_axis")
	}

	min, ok := y["min"].(float64)
	if !ok {
		return nil, fmt.Errorf("missing required attribute 'min' for y_axis")
	}

	yAxis := &client.YAxis{
		Min: min,
		Max: max,
	}

	return yAxis, nil
}

func buildTemplateVariables(templateVariablesIn []interface{}) []client.TemplateVariable {
	var newTemplateVariables []client.TemplateVariable
	for _, tv := range templateVariablesIn {
		tvMap := tv.(map[string]interface{})
		name := tvMap["name"].(string)
		suggestionAttributeKey := tvMap["suggestion_attribute_key"].(string)
		defaultValues := buildDefaultValues(tvMap["default_values"].([]interface{}))

		newTemplateVariables = append(newTemplateVariables, client.TemplateVariable{
			Name:                   name,
			DefaultValues:          defaultValues,
			SuggestionAttributeKey: suggestionAttributeKey,
		})
	}
	return newTemplateVariables
}

func buildDefaultValues(valuesIn []interface{}) []string {
	defaultValues := make([]string, 0, len(valuesIn))
	for _, v := range valuesIn {
		defaultValues = append(defaultValues, v.(string))
	}
	return defaultValues
}

func (p *resourceUnifiedDashboardImp) setResourceDataFromUnifiedDashboard(project string, dash client.UnifiedDashboard, d *schema.ResourceData, hasLegacyChartsIn bool) error {
	if err := d.Set("project_name", project); err != nil {
		return fmt.Errorf("unable to set project_name resource field: %v", err)
	}

	if err := d.Set("dashboard_name", dash.Attributes.Name); err != nil {
		return fmt.Errorf("unable to set dashboard_name resource field: %v", err)
	}

	if err := d.Set("dashboard_description", dash.Attributes.Description); err != nil {
		return fmt.Errorf("unable to set dashboard_description resource field: %v", err)
	}

	if err := d.Set("type", dash.Type); err != nil {
		return fmt.Errorf("unable to set type resource field: %v", err)
	}

	if isLegacyImplicitGroup(dash.Attributes.Groups, hasLegacyChartsIn) {
		charts, textPanels, err := assembleDashboardPanels(dash.ID, p.chartSchemaType, dash.Attributes.Groups[0].Charts)
		if err != nil {
			return err
		}
		if len(textPanels) > 0 {
			return fmt.Errorf("text panels are only supported within groups")
		}
		if err := d.Set("chart", charts); err != nil {
			return err
		}
	} else {
		var groups []interface{}
		for _, g := range dash.Attributes.Groups {
			group := map[string]interface{}{}
			group["title"] = g.Title
			group["id"] = g.ID
			group["visibility_type"] = g.VisibilityType
			group["rank"] = g.Rank

			groupCharts, groupTextPanels, err := assembleDashboardPanels(dash.ID, p.chartSchemaType, g.Charts)
			if err != nil {
				return err
			}
			group["chart"] = groupCharts
			group["text_panel"] = groupTextPanels

			groups = append(groups, group)
		}
		if err := d.Set("group", groups); err != nil {
			return fmt.Errorf("unable to set group resource field: %v", err)
		}
	}

	labels := extractLabels(dash.Attributes.Labels)
	if err := d.Set("label", labels); err != nil {
		return fmt.Errorf("unable to set labels resource field: %v", err)
	}

	var templateVariables []interface{}
	for _, tv := range dash.Attributes.TemplateVariables {
		templateVariable := map[string]interface{}{}
		templateVariable["name"] = tv.Name
		templateVariable["default_values"] = tv.DefaultValues
		templateVariable["suggestion_attribute_key"] = tv.SuggestionAttributeKey

		templateVariables = append(templateVariables, templateVariable)
	}
	if err := d.Set("template_variable", templateVariables); err != nil {
		return fmt.Errorf("unable to set template variables resource field: %v", err)
	}

	return nil
}

// assembleDashboardPanels takes the incoming set of UnifiedCharts which contain a mix
// of charts and text panels, then partitions them into separate slices to
// be processed into distinct Terraform resources.
func assembleDashboardPanels(
	dashboardID string,
	chartSchemaType ChartSchemaType,
	panels []client.UnifiedChart,
) (
	chartResources []interface{},
	textPanelResources []interface{},
	err error,
) {
	charts := []client.UnifiedChart{}
	textPanels := []client.UnifiedChart{}

	// Partition by type
	for _, panel := range panels {
		switch panel.ChartType {
		case "timeseries":
			charts = append(charts, panel)
		case "text":
			textPanels = append(textPanels, panel)
		default:
			return nil, nil, fmt.Errorf("unknown panel type: %s", panel.ChartType)
		}
	}

	chartResources, err = assembleCharts(dashboardID, chartSchemaType, charts)
	if err != nil {
		return nil, nil, err
	}
	textPanelResources, err = assembleTextPanels(dashboardID, textPanels)
	if err != nil {
		return nil, nil, err
	}

	return chartResources, textPanelResources, nil
}

// assemblePanel copies the data common to all panels (both charts and text panels)
// into the resource interface
func setPanelResourceData(
	resource map[string]interface{}, // Terraform resource
	panel client.UnifiedChart, // Panel from the API
) {
	resource["name"] = panel.Title
	resource["description"] = panel.Description
	resource["x_pos"] = panel.Position.XPos
	resource["y_pos"] = panel.Position.YPos
	resource["width"] = panel.Position.Width
	resource["height"] = panel.Position.Height

	resource["id"] = panel.ID
}

// assembleTextPanels copies the data from the API into the Terraform resource.
// Note that in Terraform charts and text panels are separate resources, but in
// the API they are both UnifiedCharts.
func assembleTextPanels(
	dashboardID string,
	textPanels []client.UnifiedChart,
) ([]interface{}, error) {
	panelResources := make([]interface{}, 0, len(textPanels))
	for _, textPanel := range textPanels {
		if textPanel.ChartType != "text" {
			return nil, fmt.Errorf("panel %s is not a text panel", textPanel.Title)
		}

		resource := map[string]interface{}{}
		setPanelResourceData(resource, textPanel)
		resource["text"] = textPanel.Text
		panelResources = append(panelResources, resource)
	}
	return panelResources, nil
}

func assembleCharts(
	dashboardID string,
	chartSchemaType ChartSchemaType,
	chartsIn []client.UnifiedChart,
) ([]interface{}, error) {
	var chartResources []interface{}
	for _, c := range chartsIn {
		if c.ChartType != "timeseries" {
			return nil, fmt.Errorf("panel %s is not a timeseries chart", c.Title)
		}

		resource := map[string]interface{}{}
		setPanelResourceData(resource, c)
		resource["rank"] = c.Rank
		resource["type"] = c.ChartType

		if c.YAxis != nil {
			resource["y_axis"] = []map[string]interface{}{
				{
					"max": c.YAxis.Max,
					"min": c.YAxis.Min,
				},
			}
		}

		if c.Subtitle != nil {
			resource["subtitle"] = *c.Subtitle
		}

		if chartSchemaType == MetricChartSchema {
			resource["query"] = getQueriesFromMetricConditionData(c.MetricQueries)
		} else {
			queries, err := getQueriesFromUnifiedDashboardResourceData(
				c.MetricQueries,
				dashboardID,
				c.ID,
			)
			if err != nil {
				return nil, err
			}
			resource["query"] = queries
		}

		chartResources = append(chartResources, resource)
	}
	return chartResources, nil
}

// isLegacyImplicitGroup defines the logic for determining if the charts in this dashboard need to be unwrapped to
// maintain backwards compatibility with the pre group definition
func isLegacyImplicitGroup(groups []client.UnifiedGroup, hasLegacyChartsIn bool) bool {
	if len(groups) != 1 {
		return false
	}
	if groups[0].VisibilityType != "implicit" {
		return false
	}
	if !hasLegacyChartsIn {
		return false
	}
	for _, c := range groups[0].Charts {
		pos := c.Position
		if pos.XPos != 0 || pos.YPos != 0 || pos.Width != 0 || pos.Height != 0 {
			return false
		}
	}
	return true
}

func (p *resourceUnifiedDashboardImp) resourceUnifiedDashboardUpdate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	c := m.(*client.Client)
	attrs, _, err := getUnifiedDashboardAttributesFromResource(d)
	if err != nil {
		return diag.FromErr(fmt.Errorf("failed to get dashboard attributes from resource : %v", err))
	}

	if _, err := c.UpdateUnifiedDashboard(ctx, d.Get("project_name").(string), d.Id(), *attrs); err != nil {
		return diag.FromErr(fmt.Errorf("failed to update dashboard: %v", err))
	}

	return p.resourceUnifiedDashboardRead(ctx, d, m)
}

func (*resourceUnifiedDashboardImp) resourceUnifiedDashboardDelete(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	var diags diag.Diagnostics

	c := m.(*client.Client)
	if err := c.DeleteUnifiedDashboard(ctx, d.Get("project_name").(string), d.Id()); err != nil {
		return diag.FromErr(fmt.Errorf("failed to delete dashboard: %v", err))
	}

	// d.SetId("") is automatically called assuming delete returns no errors, but
	// it is added here for explicitness.
	d.SetId("")
	return diags
}

func (p *resourceUnifiedDashboardImp) resourceUnifiedDashboardImport(ctx context.Context, d *schema.ResourceData, m interface{}) ([]*schema.ResourceData, error) {
	c := m.(*client.Client)

	ids := strings.Split(d.Id(), ".")
	if len(ids) != 2 {
		resourceName := "lighstep_dashboard"
		if p.chartSchemaType == MetricChartSchema {
			resourceName = "lightstep_metric_dashboard"
		}
		return []*schema.ResourceData{}, fmt.Errorf("error importing %v. Expecting an  ID formed as '<lightstep_project>.<%v_ID>'", resourceName, resourceName)
	}

	project, id := ids[0], ids[1]
	dash, err := c.GetUnifiedDashboard(ctx, project, id)
	if err != nil {
		return []*schema.ResourceData{}, fmt.Errorf("failed to get dashboard. err: %v", err)
	}
	d.SetId(id)
	if err := p.setResourceDataFromUnifiedDashboard(project, *dash, d, false); err != nil {
		return nil, fmt.Errorf("failed to set dashboard from API response to terraform state: %v", err)
	}

	return []*schema.ResourceData{d}, nil

}

// buildLabels transforms labels from the TF resource into labels for the API request
func buildLabels(labelsIn []interface{}) ([]client.Label, error) {
	var labels []client.Label

	for _, l := range labelsIn {
		label, ok := l.(map[string]interface{})
		if !ok {
			return nil, fmt.Errorf("bad format, %v", l)
		}

		if len(label) == 0 {
			continue
		}

		// label keys can be omitted for labels without the key:value syntax
		k := label["key"]
		if k == nil {
			k = ""
		}

		key, ok := k.(string)
		if !ok {
			return nil, fmt.Errorf("label key must be a string, %v", k)
		}

		v, ok := label["value"].(string)
		if !ok {
			return nil, fmt.Errorf("label value is a required field, %v", v)
		}

		labels = append(labels, client.Label{
			Key:   key,
			Value: v,
		})
	}

	return labels, nil
}

// extractLabels transforms labels from the API call into TF resource labels
func extractLabels(apiLabels []client.Label) []interface{} {
	var labels []interface{}
	for _, l := range apiLabels {
		label := map[string]interface{}{}
		if l.Key != "" {
			label["key"] = l.Key
		}
		label["value"] = l.Value
		labels = append(labels, label)
	}
	return labels
}
