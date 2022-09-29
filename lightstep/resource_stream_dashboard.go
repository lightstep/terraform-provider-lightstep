package lightstep

import (
	"context"
	"fmt"
	"net/http"
	"strings"

	"github.com/lightstep/terraform-provider-lightstep/client"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func resourceStreamDashboard() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceStreamDashboardCreate,
		ReadContext:   resourceStreamDashboardRead,
		DeleteContext: resourceStreamDashboardDelete,
		UpdateContext: resourceStreamDashboardUpdate,
		Importer: &schema.ResourceImporter{
			StateContext: resourceStreamDashboardImport,
		},
		DeprecationMessage: "resource_stream_dashboard is no longer supported. Please migrate to resource_metric_dashboard with span queries.",
		Schema: map[string]*schema.Schema{
			"dashboard_name": {
				Type:     schema.TypeString,
				Required: true,
			},
			"project_name": {
				Type:     schema.TypeString,
				Required: true,
			},
			"stream_ids": {
				Type:     schema.TypeList,
				Optional: true,
				Elem: &schema.Schema{
					Type: schema.TypeString,
				},
			},
		},
	}
}

func resourceStreamDashboardCreate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	client := m.(*client.Client)

	projectName := d.Get("project_name").(string)
	dashboardName := d.Get("dashboard_name").(string)
	streams := streamIDsToStreams(d.Get("stream_ids").([]interface{}))

	dashboard, err := client.CreateDashboard(ctx, projectName, dashboardName, streams)
	if err != nil {
		return diag.FromErr(fmt.Errorf("Failed to create stream dashboard for [project: %v; dashboard: %v]: %v", projectName, dashboardName, err))
	}

	d.SetId(dashboard.ID)
	return resourceStreamDashboardRead(ctx, d, m)
}

func resourceStreamDashboardRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	var diags diag.Diagnostics

	c := m.(*client.Client)

	projectName := d.Get("project_name").(string)
	resourceId := d.Id()

	dashboard, err := c.GetDashboard(ctx, projectName, resourceId)
	if err != nil {
		apiErr := err.(client.APIResponseCarrier)
		if apiErr.GetHTTPResponse().StatusCode == http.StatusNotFound {
			d.SetId("")
			return diags
		}
		return diag.FromErr(fmt.Errorf("Failed to get stream dashboard for [project: %v; resource_id: %v]: %v\n", projectName, resourceId, apiErr))
	}

	if err := setResourceDataFromStreamDashboard(d, *dashboard); err != nil {
		return diag.FromErr(fmt.Errorf("Failed to set stream dashboard response from API to terraform state for [project: %v; resource_id: %v]: %v", projectName, resourceId, err))
	}

	return diags
}

func resourceStreamDashboardUpdate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	var diags diag.Diagnostics

	client := m.(*client.Client)
	projectName := d.Get("project_name").(string)
	dashboardName := d.Get("dashboard_name").(string)
	resourceId := d.Id()
	streams := streamIDsToStreams(d.Get("stream_ids").([]interface{}))

	if _, err := client.UpdateDashboard(ctx, projectName, dashboardName, streams, resourceId); err != nil {
		return diag.FromErr(fmt.Errorf("Failed to update stream condition for [project: %v; dashboard_name: %v, resource_id: %v]: %v", projectName, dashboardName, resourceId, err))
	}

	return diags
}

func resourceStreamDashboardDelete(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	var diags diag.Diagnostics

	client := m.(*client.Client)
	projectName := d.Get("project_name").(string)
	resourceId := d.Id()

	if err := client.DeleteDashboard(ctx, projectName, resourceId); err != nil {
		return diag.FromErr(fmt.Errorf("Failed to delete stream dashboard for [project: %v; resource_id: %v]: %v", projectName, resourceId, err))
	}

	d.SetId("")
	return diags
}

func resourceStreamDashboardImport(ctx context.Context, d *schema.ResourceData, m interface{}) ([]*schema.ResourceData, error) {
	client := m.(*client.Client)

	resourceId := d.Id()
	ids := strings.Split(resourceId, ".")
	if len(ids) != 2 {
		return []*schema.ResourceData{}, fmt.Errorf("error importing lightstep_dashboard. Expecting an  ID formed as '<lightstep_project>.<lightstep_dashboardID>' (provided: %v)", resourceId)
	}
	project, id := ids[0], ids[1]

	dashboard, err := client.GetDashboard(ctx, project, id)
	if err != nil {
		return []*schema.ResourceData{}, err
	}

	d.SetId(id)
	if err := d.Set("project_name", project); err != nil {
		return []*schema.ResourceData{}, err
	}

	if err := setResourceDataFromStreamDashboard(d, *dashboard); err != nil {
		return []*schema.ResourceData{}, fmt.Errorf("Failed to set stream dashboard from API response to terraform state: %v", err)
	}

	return []*schema.ResourceData{d}, nil
}

func setResourceDataFromStreamDashboard(d *schema.ResourceData, dashboard client.Dashboard) error {
	if err := d.Set("dashboard_name", dashboard.Attributes.Name); err != nil {
		return fmt.Errorf("Unable to set dashboard_name resource field: %v", err)
	}

	var streamIDs []string
	for _, stream := range dashboard.Attributes.Streams {
		streamIDs = append(streamIDs, stream.ID)
	}

	if err := d.Set("stream_ids", streamIDs); err != nil {
		return fmt.Errorf("Unable to set stream_ids resource field: %v", err)
	}

	return nil
}

func streamIDsToStreams(ids []interface{}) []client.Stream {
	streams := []client.Stream{}

	for _, id := range ids {
		streams = append(streams, client.Stream{ID: id.(string)})
	}
	return streams
}
