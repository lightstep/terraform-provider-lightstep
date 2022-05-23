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

func resourceMetricDashboard() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceMetricDashboardCreate,
		ReadContext:   resourceMetricDashboardRead,
		UpdateContext: resourceMetricDashboardUpdate,
		DeleteContext: resourceMetricDashboardDelete,
		Importer: &schema.ResourceImporter{
			StateContext: resourceMetricDashboardImport,
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
			"type": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"chart": {
				Type:     schema.TypeSet,
				Optional: true,
				Elem: &schema.Resource{
					Schema: getChartSchema(),
				},
			},
		},
	}
}

func getChartSchema() map[string]*schema.Schema {
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
				Schema: getQuerySchema(true),
			},
		},
	}
}

func resourceMetricDashboardCreate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	c := m.(*client.Client)
	attrs, err := getMetricDashboardAttributesFromResource(d)
	if err != nil {
		return diag.FromErr(fmt.Errorf("Failed to get metric dashboard attributes: %v", err))
	}

	dashboard := client.MetricDashboard{
		Type:       "dashboard",
		Attributes: *attrs,
	}

	created, err := c.CreateMetricDashboard(ctx, d.Get("project_name").(string), dashboard)
	if err != nil {
		return diag.FromErr(fmt.Errorf("Failed to create metric dashboard: %v", err))
	}

	d.SetId(created.ID)
	return resourceMetricDashboardRead(ctx, d, m)
}

func resourceMetricDashboardRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	var diags diag.Diagnostics

	c := m.(*client.Client)

	dashboard, err := c.GetMetricDashboard(ctx, d.Get("project_name").(string), d.Id())
	if err != nil {
		apiErr := err.(client.APIResponseCarrier)
		if apiErr.GetHTTPResponse() != nil && apiErr.GetHTTPResponse().StatusCode == http.StatusNotFound {
			d.SetId("")
			return diags
		}
		return diag.FromErr(fmt.Errorf("Failed to get metric dashboard: %v\n", apiErr))
	}

	if err := setResourceDataFromMetricDashboard(d.Get("project_name").(string), *dashboard, d); err != nil {
		return diag.FromErr(fmt.Errorf("Failed to set metric dashboard from API response to terraform state: %v", err))
	}

	return diags
}

func getMetricDashboardAttributesFromResource(d *schema.ResourceData) (*client.MetricDashboardAttributes, error) {
	chartSet := d.Get("chart").(*schema.Set)
	charts, err := buildCharts(chartSet.List())
	if err != nil {
		return nil, err
	}

	attributes := &client.MetricDashboardAttributes{
		Name:   d.Get("dashboard_name").(string),
		Charts: charts,
	}

	return attributes, nil
}

func buildCharts(chartsIn []interface{}) ([]client.MetricChart, error) {
	var (
		charts    []map[string]interface{}
		newCharts []client.MetricChart
	)

	for _, chart := range chartsIn {
		charts = append(charts, chart.(map[string]interface{}))
	}

	for _, chart := range charts {
		c := client.MetricChart{
			Title:     chart["name"].(string),
			Rank:      chart["rank"].(int),
			ID:        chart["id"].(string),
			ChartType: chart["type"].(string),
		}

		queries, err := buildQueries(chart["query"].([]interface{}), true)
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
		return nil, fmt.Errorf("Missing required attribute 'max' for y_axis")
	}

	min, ok := y["min"].(float64)
	if !ok {
		return nil, fmt.Errorf("Missing required attribute 'min' for y_axis")
	}

	yAxis := &client.YAxis{
		Min: min,
		Max: max,
	}

	return yAxis, nil
}

func setResourceDataFromMetricDashboard(project string, dash client.MetricDashboard, d *schema.ResourceData) error {
	if err := d.Set("project_name", project); err != nil {
		return fmt.Errorf("Unable to set project_name resource field: %v", err)
	}

	if err := d.Set("dashboard_name", dash.Attributes.Name); err != nil {
		return fmt.Errorf("Unable to set dashboard_name resource field: %v", err)
	}

	if err := d.Set("type", dash.Type); err != nil {
		return fmt.Errorf("Unable to set type resource field: %v", err)
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

		chart["query"] = getQueriesFromResourceData(c.MetricQueries, true)
		chart["name"] = c.Title
		chart["rank"] = c.Rank
		chart["type"] = c.ChartType
		chart["id"] = c.ID

		charts = append(charts, chart)
	}

	if err := d.Set("chart", charts); err != nil {
		return fmt.Errorf("Unable to set chart resource field: %v", err)
	}

	return nil
}

func resourceMetricDashboardUpdate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	c := m.(*client.Client)
	attrs, err := getMetricDashboardAttributesFromResource(d)
	if err != nil {
		return diag.FromErr(fmt.Errorf("Failed to get metric dashboard attributes from resource : %v", err))
	}

	if _, err := c.UpdateMetricDashboard(ctx, d.Get("project_name").(string), d.Id(), *attrs); err != nil {
		return diag.FromErr(fmt.Errorf("Failed to update metric dashboard: %v", err))
	}

	return resourceMetricDashboardRead(ctx, d, m)
}

func resourceMetricDashboardDelete(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	var diags diag.Diagnostics

	c := m.(*client.Client)
	if err := c.DeleteMetricDashboard(ctx, d.Get("project_name").(string), d.Id()); err != nil {
		return diag.FromErr(fmt.Errorf("Failed to detele metrics dashboard: %v", err))
	}

	// d.SetId("") is automatically called assuming delete returns no errors, but
	// it is added here for explicitness.
	d.SetId("")
	return diags
}

func resourceMetricDashboardImport(ctx context.Context, d *schema.ResourceData, m interface{}) ([]*schema.ResourceData, error) {
	c := m.(*client.Client)

	ids := strings.Split(d.Id(), ".")
	if len(ids) != 2 {
		return []*schema.ResourceData{}, fmt.Errorf("Error importing lightstep_metric_dashboard. Expecting an  ID formed as '<lightstep_project>.<lightstep_metric_dashboard_ID>'")
	}

	project, id := ids[0], ids[1]
	dash, err := c.GetMetricDashboard(ctx, project, id)
	if err != nil {
		return []*schema.ResourceData{}, fmt.Errorf("Failed to get metric dashboard. err: %v", err)
	}
	d.SetId(id)
	if err := setResourceDataFromMetricDashboard(project, *dash, d); err != nil {
		return nil, fmt.Errorf("Failed to set metric dashboard from API response to terraform state: %v", err)
	}

	return []*schema.ResourceData{d}, nil

}
