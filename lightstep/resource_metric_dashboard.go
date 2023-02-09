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

	if err := p.setResourceDataFromUnifiedDashboard(d.Get("project_name").(string), *dashboard, d); err != nil {
		return diag.FromErr(fmt.Errorf("failed to set dashboard from API response to terraform state: %v", err))
	}

	return diags
}

func getUnifiedDashboardAttributesFromResource(d *schema.ResourceData) (*client.UnifiedDashboardAttributes, error) {
	chartSet := d.Get("chart").(*schema.Set)
	charts, err := buildCharts(chartSet.List())
	if err != nil {
		return nil, err
	}

	attributes := &client.UnifiedDashboardAttributes{
		Name:        d.Get("dashboard_name").(string),
		Description: d.Get("dashboard_description").(string),
		Charts:      charts,
	}

	return attributes, nil
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
			Title:     chart["name"].(string),
			Rank:      chart["rank"].(int),
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

	var charts []interface{}
	for _, c := range dash.Attributes.Charts {
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
				return err
			}
			chart["query"] = queries
		}
		chart["name"] = c.Title
		chart["rank"] = c.Rank
		chart["type"] = c.ChartType
		chart["id"] = c.ID

		charts = append(charts, chart)
	}

	if err := d.Set("chart", charts); err != nil {
		return fmt.Errorf("unable to set chart resource field: %v", err)
	}

	return nil
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
