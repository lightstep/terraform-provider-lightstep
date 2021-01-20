package main

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/lightstep/terraform-provider-lightstep/lightstep"
)

func resourceMetricDashboard() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceMetricDashboardCreate,
		ReadContext:   resourceMetricDashboardRead,
		UpdateContext: resourceMetricDashboardUpdate,
		DeleteContext: resourceMetricDashboardDelete,
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
				Type:     schema.TypeList,
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
			Type:     schema.TypeInt,
			Required: true,
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
			Type:     schema.TypeList,
			MaxItems: 1,
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
				Schema: getQuerySchema(),
			},
		},
	}
}

func resourceMetricDashboardCreate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	client := m.(*lightstep.Client)
	attrs, err := getMetricDashboardAttributesFromResource(d)
	if err != nil {
		return diag.FromErr(fmt.Errorf("Failed to get metric dashboard attributes: %v", err))
	}

	dashboard := lightstep.MetricDashboard{
		Type:       "dashboard",
		Attributes: *attrs,
	}

	created, err := client.CreateMetricDashboard(d.Get("project_name").(string), dashboard)
	if err != nil {
		return diag.FromErr(fmt.Errorf("Failed to create metric dashboard: %v", err))
	}

	d.SetId(created.ID)
	return resourceMetricDashboardRead(ctx, d, m)
}

func resourceMetricDashboardRead(_ context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	var diags diag.Diagnostics

	client := m.(*lightstep.Client)

	dashboard, err := client.GetMetricDashboard(d.Get("project_name").(string), d.Id())
	if err != nil {
		return diag.FromErr(fmt.Errorf("Failed to get metric dashboard: %v", err))

	}

	if err := setResourceDataFromMetricDashboard(d.Get("project_name").(string), dashboard, d); err != nil {
		return diag.FromErr(fmt.Errorf("Failed to set metric dashboard from API response to terraform state: %v", err))
	}

	return diags
}

func getMetricDashboardAttributesFromResource(d *schema.ResourceData) (*lightstep.MetricDashboardAttributes, error) {
	charts, err := buildCharts(d.Get("chart").([]interface{}))
	if err != nil {
		return nil, err
	}

	attributes := &lightstep.MetricDashboardAttributes{
		Name:   d.Get("dashboard_name").(string),
		Charts: charts,
	}

	return attributes, nil
}

func buildCharts(chartsIn []interface{}) ([]lightstep.MetricChart, error) {
	var (
		charts    []map[string]interface{}
		newCharts []lightstep.MetricChart
	)

	for _, chart := range chartsIn {
		charts = append(charts, chart.(map[string]interface{}))
	}

	for _, chart := range charts {
		c := lightstep.MetricChart{
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
		c.YAxis = yaxis
		newCharts = append(newCharts, c)
	}
	return newCharts, nil
}

func buildYAxis(yAxisIn []interface{}) (*lightstep.YAxis, error) {
	y := yAxisIn[0].(map[string]interface{})

	max, ok := y["max"].(float64)
	if !ok {
		return nil, fmt.Errorf("Missing required attribute 'max' for y_axis")
	}

	min, ok := y["min"].(float64)
	if !ok {
		return nil, fmt.Errorf("Missing required attribute 'min' for y_axis")
	}

	yAxis := &lightstep.YAxis{
		Min: min,
		Max: max,
	}

	return yAxis, nil
}

func setResourceDataFromMetricDashboard(project string, dash lightstep.MetricDashboard, d *schema.ResourceData) error {
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
		yMap := map[string]interface{}{}

		if c.YAxis != nil {
			yMap["max"] = c.YAxis.Max
			yMap["min"] = c.YAxis.Min
		}

		queries := getQueriesFromResourceData(c.MetricQueries)

		charts = append(charts, map[string]interface{}{
			"name":   c.Title,
			"rank":   c.Rank,
			"type":   c.ChartType,
			"id":     c.ID,
			"y_axis": []interface{}{yMap},
			"query":  queries,
		})
	}

	if err := d.Set("chart", charts); err != nil {
		return fmt.Errorf("Unable to set chart resource field: %v", err)
	}

	return nil
}

func resourceMetricDashboardUpdate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	client := m.(*lightstep.Client)
	attrs, err := getMetricDashboardAttributesFromResource(d)
	if err != nil {
		return diag.FromErr(fmt.Errorf("Failed to get metric dashboard attributes from resource : %v", err))
	}

	if _, err := client.UpdateMetricDashboard(d.Get("project_name").(string), d.Id(), *attrs); err != nil {
		return diag.FromErr(fmt.Errorf("Failed to update metric dashboard: %v", err))
	}

	return resourceMetricDashboardRead(ctx, d, m)
}

func resourceMetricDashboardDelete(_ context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	var diags diag.Diagnostics

	client := m.(*lightstep.Client)
	if err := client.DeleteMetricDashboard(d.Get("project_name").(string), d.Id()); err != nil {
		return diag.FromErr(fmt.Errorf("Failed to detele metrics dashboard: %v", err))
	}

	// d.SetId("") is automatically called assuming delete returns no errors, but
	// it is added here for explicitness.
	d.SetId("")
	return diags
}
