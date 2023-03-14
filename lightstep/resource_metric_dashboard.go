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
			Elem: &schema.Resource{
				Schema: getChartSchema(chartSchemaType),
			},
		},
	}
}

func getChartSchema(chartSchemaType ChartSchemaType) map[string]*schema.Schema {

	var querySchema map[string]*schema.Schema
	if chartSchemaType == UnifiedChartSchema {
		querySchema = getUnifiedQuerySchema()
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
		querySchema = getMetricQuerySchema()
	}

	return map[string]*schema.Schema{
		"name": {
			Type:     schema.TypeString,
			Required: true,
		},
		"rank": {
			Type:         schema.TypeInt,
			ValidateFunc: validation.IntAtLeast(0),
			Required:     true,
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
		"type": {
			Type:         schema.TypeString,
			Required:     true,
			ValidateFunc: validation.StringInSlice([]string{"timeseries"}, true),
		},
		"id": {
			Type:     schema.TypeString,
			Computed: true,
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
	}
}

type resourceUnifiedDashboardImp struct {
	chartSchemaType ChartSchemaType
}

func (p *resourceUnifiedDashboardImp) resourceUnifiedDashboardCreate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	c := m.(*client.Client)
	attrs, err := getUnifiedDashboardAttributesFromResource(d)
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
		if err := p.setResourceDataFromUnifiedDashboard(projectName, dashboard, d); err != nil {
			return diag.FromErr(fmt.Errorf("failed to set dashboard from API response to terraform state: %v", err))
		}
		return nil
	}

	return p.resourceUnifiedDashboardRead(ctx, d, m)
}

func (p *resourceUnifiedDashboardImp) resourceUnifiedDashboardRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	c := m.(*client.Client)

	// The lightstep_dashboard resource always wants to use query_strings rather than
	// JSON-based queries
	convertToQueryString := false
	if p.chartSchemaType == UnifiedChartSchema {
		convertToQueryString = true
	}

	prevAttrs, err := getUnifiedDashboardAttributesFromResource(d)
	if err != nil {
		return diag.FromErr(fmt.Errorf("failed to translate resource attributes: %v", err))
	}

	dashboard, err := c.GetUnifiedDashboard(ctx, d.Get("project_name").(string), d.Id(), convertToQueryString)
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

	if err := p.setResourceDataFromUnifiedDashboard(d.Get("project_name").(string), *dashboard, d); err != nil {
		return diag.FromErr(fmt.Errorf("failed to set dashboard from API response to terraform state: %v", err))
	}
	return diags
}

func getUnifiedDashboardAttributesFromResource(d *schema.ResourceData) (*client.UnifiedDashboardAttributes, error) {
	chartSet := d.Get("chart").(*schema.Set)
	groupSet := d.Get("group").(*schema.Set)
	groups, err := buildGroups(groupSet.List(), chartSet.List())
	if err != nil {
		return nil, err
	}

	labelSet := d.Get("label").(*schema.Set)
	labels, err := buildLabels(labelSet.List())
	if err != nil {
		return nil, err
	}

	attributes := &client.UnifiedDashboardAttributes{
		Name:        d.Get("dashboard_name").(string),
		Description: d.Get("dashboard_description").(string),
		Groups:      groups,
		Labels:      labels,
	}

	return attributes, nil
}

func buildGroups(groupsIn []interface{}, legacyChartsIn []interface{}) ([]client.UnifiedGroup, error) {
	var (
		groups    []map[string]interface{}
		newGroups []client.UnifiedGroup
	)

	if len(legacyChartsIn) != 0 {
		c, err := buildCharts(legacyChartsIn)
		if err != nil {
			return nil, err
		}
		newGroups = append(newGroups, client.UnifiedGroup{
			Rank:           0,
			Title:          "",
			VisibilityType: "implicit",
			Charts:         c,
		})
	}

	for _, group := range groupsIn {
		groups = append(groups, group.(map[string]interface{}))
	}

	for _, group := range groups {
		c, err := buildCharts(group["chart"].(*schema.Set).List())
		if err != nil {
			return nil, err
		}
		g := client.UnifiedGroup{
			ID:             group["id"].(string),
			Rank:           group["rank"].(int),
			Title:          group["title"].(string),
			VisibilityType: group["visibility_type"].(string),
			Charts:         c,
		}
		newGroups = append(newGroups, g)
	}
	return newGroups, nil
}

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

func buildCharts(chartsIn []interface{}) ([]client.UnifiedChart, error) {
	var (
		charts    []map[string]interface{}
		newCharts []client.UnifiedChart
	)

	for _, chart := range chartsIn {
		charts = append(charts, chart.(map[string]interface{}))
	}

	for _, chart := range charts {
		p := client.UnifiedPosition{
			XPos:   chart["x_pos"].(int),
			YPos:   chart["y_pos"].(int),
			Width:  chart["width"].(int),
			Height: chart["height"].(int),
		}
		c := client.UnifiedChart{
			Title:     chart["name"].(string),
			Rank:      chart["rank"].(int),
			Position:  p,
			ID:        chart["id"].(string),
			ChartType: chart["type"].(string),
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

func (p *resourceUnifiedDashboardImp) setResourceDataFromUnifiedDashboard(project string, dash client.UnifiedDashboard, d *schema.ResourceData) error {
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

	assembleCharts := func(chartsIn []client.UnifiedChart) ([]interface{}, error) {
		var charts []interface{}
		for _, c := range chartsIn {
			chart := map[string]interface{}{}

			yMap := map[string]interface{}{}

			if c.YAxis != nil {
				yMap["max"] = c.YAxis.Max
				yMap["min"] = c.YAxis.Min
				chart["y_axis"] = []map[string]interface{}{yMap}
			}

			if p.chartSchemaType == MetricChartSchema {
				chart["query"] = getQueriesFromMetricConditionData(c.MetricQueries)
			} else {
				queries, err := getQueriesFromUnifiedDashboardResourceData(
					c.MetricQueries,
					dash.ID,
					c.ID,
				)
				if err != nil {
					return nil, err
				}
				chart["query"] = queries
			}
			chart["name"] = c.Title
			chart["rank"] = c.Rank
			chart["x_pos"] = c.Position.XPos
			chart["y_pos"] = c.Position.YPos
			chart["width"] = c.Position.Width
			chart["height"] = c.Position.Height
			chart["type"] = c.ChartType
			chart["id"] = c.ID

			charts = append(charts, chart)
		}
		return charts, nil
	}
	if isLegacyImplicitGroup(dash.Attributes.Groups) {
		charts, err := assembleCharts(dash.Attributes.Groups[0].Charts)
		if err != nil {
			return err
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

			groupCharts, err := assembleCharts(g.Charts)
			if err != nil {
				return err
			}
			group["chart"] = groupCharts
			groups = append(groups, group)
		}
		if err := d.Set("group", groups); err != nil {
			return fmt.Errorf("unable to set group resource field: %v", err)
		}
	}

	var labels []interface{}
	for _, l := range dash.Attributes.Labels {
		label := map[string]interface{}{}
		if l.Key != "" {
			label["key"] = l.Key
		}
		label["value"] = l.Value
		labels = append(labels, label)
	}
	if err := d.Set("label", labels); err != nil {
		return fmt.Errorf("unable to set labels resource field: %v", err)
	}

	return nil
}

// isLegacyImplicitGroup defines the logic for determining if the charts in this dashboard need to be unwrapped to
// maintain backwards compatibility with the pre group definition
func isLegacyImplicitGroup(groups []client.UnifiedGroup) bool {
	if len(groups) != 1 {
		return false
	}
	if groups[0].VisibilityType != "implicit" {
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
	attrs, err := getUnifiedDashboardAttributesFromResource(d)
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

	convertToQueryString := false
	if p.chartSchemaType == UnifiedChartSchema {
		convertToQueryString = true
	}

	project, id := ids[0], ids[1]
	dash, err := c.GetUnifiedDashboard(ctx, project, id, convertToQueryString)
	if err != nil {
		return []*schema.ResourceData{}, fmt.Errorf("failed to get dashboard. err: %v", err)
	}
	d.SetId(id)
	if err := p.setResourceDataFromUnifiedDashboard(project, *dash, d); err != nil {
		return nil, fmt.Errorf("failed to set dashboard from API response to terraform state: %v", err)
	}

	return []*schema.ResourceData{d}, nil

}
