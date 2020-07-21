package main

import (
	"fmt"
	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
	"github.com/lightstep/terraform-provider-lightstep/lightstep"
	"strings"
)

func resourceDashboard() *schema.Resource {
	return &schema.Resource{
		Create: resourceDashboardCreate,
		Read:   resourceDashboardRead,
		Delete: resourceDashboardDelete,
		Exists: resourceDashboardExists,
		Update: resourceDashboardUpdate,
		Importer: &schema.ResourceImporter{
			State: resourceDashboardImport,
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

func resourceDashboardExists(d *schema.ResourceData, m interface{}) (b bool, e error) {
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

func resourceDashboardCreate(d *schema.ResourceData, m interface{}) error {
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
	return resourceDashboardRead(d, m)
}

func resourceDashboardRead(d *schema.ResourceData, m interface{}) error {
	client := m.(*lightstep.Client)
	_, err := client.GetDashboard(
		d.Get("project_name").(string),
		d.Id(),
	)
	if err != nil {
		return err
	}
	return nil
}

func resourceDashboardUpdate(d *schema.ResourceData, m interface{}) error {
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

func resourceDashboardDelete(d *schema.ResourceData, m interface{}) error {
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

func resourceDashboardImport(d *schema.ResourceData, m interface{}) ([]*schema.ResourceData, error) {
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
	d.Set("project_name", project) // nolint  these values are fetched from LS
	d.Set("stream_ids", streamIDs) // nolint  and are known to be valid
	return []*schema.ResourceData{d}, nil
}
