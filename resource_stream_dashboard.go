package main

import (
	"context"
	"fmt"
	"net/http"
	"strings"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/lightstep/terraform-provider-lightstep/lightstep"
)

func resourceStreamDashboard() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceStreamDashboardCreate,
		ReadContext:   resourceStreamDashboardRead,
		DeleteContext: resourceStreamDashboardDelete,
		// TODO Exists is deprecated
		Exists:        resourceStreamDashboardExists,
		UpdateContext: resourceStreamDashboardUpdate,
		Importer: &schema.ResourceImporter{
			State: resourceStreamDashboardImport,
		},
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

func resourceStreamDashboardExists(d *schema.ResourceData, m interface{}) (b bool, e error) {
	client := m.(*lightstep.Client)

	if _, err := client.GetDashboard(d.Get("project_name").(string), d.Id()); err != nil {
		return false, fmt.Errorf("Failed to get stream dashboard: %v", err)
	}

	return true, nil
}

func resourceStreamDashboardCreate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	client := m.(*lightstep.Client)

	streams := streamIDsToStreams(d.Get("stream_ids").([]interface{}))
	dashboard, err := client.CreateDashboard(
		d.Get("project_name").(string),
		d.Get("dashboard_name").(string),
		streams,
	)
	if err != nil {
		return diag.FromErr(fmt.Errorf("Failed to create stream dashboard: %v", err))
	}

	d.SetId(dashboard.ID)
	return resourceStreamDashboardRead(ctx, d, m)
}

func resourceStreamDashboardRead(_ context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	var diags diag.Diagnostics

	client := m.(*lightstep.Client)
	dashboard, err := client.GetDashboard(d.Get("project_name").(string), d.Id())
	if err != nil {
		apiErr := err.(lightstep.APIResponseCarrier)
		if apiErr.GetHTTPResponse().StatusCode == http.StatusNotFound {
			d.SetId("")
			return diags
		}
		return diag.FromErr(fmt.Errorf("Failed to get stream dashboard: %v\n", apiErr))
	}

	if err := setResourceDataFromStreamDashboard(d, *dashboard); err != nil {
		return diag.FromErr(fmt.Errorf("Failed to set stream dashboard response from API to terraform state: %v", err))
	}

	return diags
}

func resourceStreamDashboardUpdate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	var diags diag.Diagnostics

	client := m.(*lightstep.Client)
	streams := streamIDsToStreams(d.Get("stream_ids").([]interface{}))

	if _, err := client.UpdateDashboard(
		d.Get("project_name").(string),
		d.Get("dashboard_name").(string),
		streams,
		d.Id(),
	); err != nil {
		return diag.FromErr(fmt.Errorf("Failed to update stream condition: %v", err))
	}

	return diags
}

func resourceStreamDashboardDelete(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	var diags diag.Diagnostics

	client := m.(*lightstep.Client)
	if err := client.DeleteDashboard(d.Get("project_name").(string), d.Id()); err != nil {
		return diag.FromErr(fmt.Errorf("Failed to delete stream dashboard: %v", err))
	}

	d.SetId("")
	return diags
}

func resourceStreamDashboardImport(d *schema.ResourceData, m interface{}) ([]*schema.ResourceData, error) {
	client := m.(*lightstep.Client)

	ids := strings.Split(d.Id(), ".")
	if len(ids) != 2 {
		return []*schema.ResourceData{}, fmt.Errorf("Error importing lightstep_dashboard. Expecting an  ID formed as '<lightstep_project>.<lightstep_dashboardID>'")
	}
	project, id := ids[0], ids[1]

	dashboard, err := client.GetDashboard(project, id)
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

func setResourceDataFromStreamDashboard(d *schema.ResourceData, dashboard lightstep.Dashboard) error {
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

func streamIDsToStreams(ids []interface{}) []lightstep.Stream {
	streams := []lightstep.Stream{}

	for _, id := range ids {
		streams = append(streams, lightstep.Stream{ID: id.(string)})
	}
	return streams
}
