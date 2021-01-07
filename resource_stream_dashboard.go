package main

import (
	"fmt"
	"strings"

	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
	"github.com/lightstep/terraform-provider-lightstep/lightstep"
)

func resourceStreamDashboard() *schema.Resource {
	return &schema.Resource{
		Create: resourceStreamDashboardCreate,
		Read:   resourceStreamDashboardRead,
		Delete: resourceStreamDashboardDelete,
		Exists: resourceStreamDashboardExists,
		Update: resourceStreamDashboardUpdate,
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
	_, err := client.GetDashboard(
		d.Get("project_name").(string),
		d.Id(),
	)
	if err != nil {
		return false, err
	}
	return true, nil
}

func resourceStreamDashboardCreate(d *schema.ResourceData, m interface{}) error {
	client := m.(*lightstep.Client)

	streams := streamIDsToStreams(d.Get("stream_ids").([]interface{}))

	dashboard, err := client.CreateDashboard(
		d.Get("project_name").(string),
		d.Get("dashboard_name").(string),
		streams,
	)
	if err != nil {
		return err
	}

	d.SetId(dashboard.ID)
	return resourceStreamDashboardRead(d, m)
}

func resourceStreamDashboardRead(d *schema.ResourceData, m interface{}) error {
	client := m.(*lightstep.Client)
	dashboard, err := client.GetDashboard(d.Get("project_name").(string), d.Id())
	if err != nil {
		return err
	}

	if err := readResourceDataFromStreamDashboard(d, dashboard); err != nil {
		return err
	}

	return nil
}

func readResourceDataFromStreamDashboard(d *schema.ResourceData, dashboard lightstep.Dashboard) error {
	if err := d.Set("dashboard_name", dashboard.Attributes.Name); err != nil {
		return err
	}

	var streamIDs []string
	for _, stream := range dashboard.Attributes.Streams {
		streamIDs = append(streamIDs, stream.ID)
	}

	if err := d.Set("stream_ids", streamIDs); err != nil {
		return err
	}

	return nil
}

func resourceStreamDashboardUpdate(d *schema.ResourceData, m interface{}) error {
	client := m.(*lightstep.Client)

	streams := streamIDsToStreams(d.Get("stream_ids").([]interface{}))

	_, err := client.UpdateDashboard(
		d.Get("project_name").(string),
		d.Get("dashboard_name").(string),
		streams,
		d.Id(),
	)
	if err != nil {
		return err
	}
	return nil
}

func resourceStreamDashboardDelete(d *schema.ResourceData, m interface{}) error {
	client := m.(*lightstep.Client)
	err := client.DeleteDashboard(
		d.Get("project_name").(string),
		d.Id(),
	)
	if err != nil {
		return err
	}
	d.SetId("")
	return nil
}

func streamIDsToStreams(ids []interface{}) []lightstep.Stream {
	streams := []lightstep.Stream{}

	for _, id := range ids {
		streams = append(streams, lightstep.Stream{ID: id.(string)})
	}
	return streams
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

	var streamIDs []string
	for _, stream := range dashboard.Attributes.Streams {
		streamIDs = append(streamIDs, stream.ID)
	}

	d.SetId(id)
	if err := d.Set("project_name", project); err != nil {
		return []*schema.ResourceData{}, err
	}

	if err := d.Set("dashboard_name", dashboard.Attributes.Name); err != nil {
		return []*schema.ResourceData{}, err
	}

	if err := d.Set("stream_ids", streamIDs); err != nil {
		return []*schema.ResourceData{d}, err
	}

	return []*schema.ResourceData{d}, nil
}
