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

func resourceDashboard() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceDashboardCreate,
		ReadContext:   resourceDashboardRead,
		UpdateContext: resourceDashboardUpdate,
		DeleteContext: resourceDashboardDelete,
		Importer: &schema.ResourceImporter{
			StateContext: resourceDashboardImport,
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
					Schema: getChartSchema(UnifiedChartSchema),
				},
			},
		},
	}
}

func resourceDashboardCreate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
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
	return resourceDashboardRead(ctx, d, m)
}

func resourceDashboardRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	var diags diag.Diagnostics

	c := m.(*client.Client)

	dashboard, err := c.GetUnifiedDashboard(ctx, d.Get("project_name").(string), d.Id())
	if err != nil {
		apiErr := err.(client.APIResponseCarrier)
		if apiErr.GetHTTPResponse() != nil && apiErr.GetHTTPResponse().StatusCode == http.StatusNotFound {
			d.SetId("")
			return diags
		}
		return diag.FromErr(fmt.Errorf("failed to get dashboard: %v", apiErr))
	}

	if err := setResourceDataFromDashboard(d.Get("project_name").(string), *dashboard, d); err != nil {
		return diag.FromErr(fmt.Errorf("failed to set dashboard from API response to terraform state: %v", err))
	}

	return diags
}

func setResourceDataFromDashboard(project string, dash client.UnifiedDashboard, d *schema.ResourceData) error {
	if err := d.Set("project_name", project); err != nil {
		return fmt.Errorf("unable to set project_name resource field: %v", err)
	}

	if err := d.Set("dashboard_name", dash.Attributes.Name); err != nil {
		return fmt.Errorf("unable to set dashboard_name resource field: %v", err)
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

		chart["query"] = getQueriesFromDashboardResourceData(c.MetricQueries)
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

func getQueriesFromDashboardResourceData(queriesIn []client.MetricQueryWithAttributes) []interface{} {
	var queries []interface{}
	for _, q := range queriesIn {
		qs := map[string]interface{}{
			"hidden":       q.Hidden,
			"display":      q.Display,
			"query_name":   q.Name,
			"query_string": q.TQLQuery,
		}
		queries = append(queries, qs)
	}
	return queries
}

func resourceDashboardUpdate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	c := m.(*client.Client)
	attrs, err := getUnifiedDashboardAttributesFromResource(d)
	if err != nil {
		return diag.FromErr(fmt.Errorf("failed to get dashboard attributes from resource : %v", err))
	}

	if _, err := c.UpdateUnifiedDashboard(ctx, d.Get("project_name").(string), d.Id(), *attrs); err != nil {
		return diag.FromErr(fmt.Errorf("failed to update dashboard: %v", err))
	}

	return resourceDashboardRead(ctx, d, m)
}

func resourceDashboardDelete(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	var diags diag.Diagnostics

	c := m.(*client.Client)
	if err := c.DeleteDashboard(ctx, d.Get("project_name").(string), d.Id()); err != nil {
		return diag.FromErr(fmt.Errorf("failed to detele dashboard: %v", err))
	}

	// d.SetId("") is automatically called assuming delete returns no errors, but
	// it is added here for explicitness.
	d.SetId("")
	return diags
}

func resourceDashboardImport(ctx context.Context, d *schema.ResourceData, m interface{}) ([]*schema.ResourceData, error) {
	c := m.(*client.Client)

	ids := strings.Split(d.Id(), ".")
	if len(ids) != 2 {
		return []*schema.ResourceData{}, fmt.Errorf("error importing lightstep_dashboard. Expecting an  ID formed as '<lightstep_project>.<lightstep_dashboard_ID>'")
	}

	project, id := ids[0], ids[1]
	dash, err := c.GetUnifiedDashboard(ctx, project, id)
	if err != nil {
		return []*schema.ResourceData{}, fmt.Errorf("failed to get dashboard. err: %v", err)
	}
	d.SetId(id)
	if err := setResourceDataFromDashboard(project, *dash, d); err != nil {
		return nil, fmt.Errorf("failed to set dashboard from API response to terraform state: %v", err)
	}

	return []*schema.ResourceData{d}, nil

}

func getUnifiedQuerySchema() map[string]*schema.Schema {
	sma := map[string]*schema.Schema{
		"hidden": {
			Type:     schema.TypeBool,
			Required: true,
		},
		"display": {
			Type:         schema.TypeString,
			Optional:     true,
			ValidateFunc: validation.StringInSlice([]string{"line", "area", "bar", "big_number", "heatmap"}, false),
		},
		"query_name": {
			Type:     schema.TypeString,
			Required: true,
		},
		"query_string": {
			Type:     schema.TypeString,
			Required: true,
		},
	}
	return sma
}
