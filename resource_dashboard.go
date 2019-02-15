package main

import (
	"github.com/hashicorp/terraform/helper/schema"
	"github.com/lightstep/terraform-provider-lightstep/lightstep"
	"log"
)

func resourceDashboard() *schema.Resource {
	return &schema.Resource{
		Create: resourceDashboardCreate,
		Read:   resourceDashboardRead,
		Delete: resourceDashboardDelete,
		Exists: resourceDashboardExists,
		Update: resourceDashboardUpdate,

		Schema: map[string]*schema.Schema{
			"dashboard_name": {
				Type:     schema.TypeString,
				Required: true,
			},
			"project_name": {
				Type:     schema.TypeString,
				Required: true,
			},
			"streams": {
				Type:     schema.TypeList,
				Required: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"stream_name": &schema.Schema{
							Type:     schema.TypeString,
							Required: true,
						},
						"query": &schema.Schema{
							Type:     schema.TypeString,
							Required: true,
						},
					},
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

	var searchAttributes []lightstep.SearchAttributes
	for _, sa := range d.Get("streams").([]interface{}) {
		sa, _ := sa.(map[string]interface{})
		searchAttributes = append(
			searchAttributes,
			lightstep.SearchAttributes{
				Name:  sa["stream_name"].(string),
				Query: sa["query"].(string),
			},
		)
	}

	resp, err := client.CreateDashboard(
		d.Get("project_name").(string),
		d.Get("dashboard_name").(string),
		searchAttributes,
	)
	if err != nil {
		log.Println(err)
		return err
	}
	d.SetId(string(resp.Data.ID))
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

	var searchAttributes []lightstep.SearchAttributes
	for _, sa := range d.Get("streams").([]interface{}) {
		sa, _ := sa.(map[string]interface{})
		searchAttributes = append(
			searchAttributes,
			lightstep.SearchAttributes{
				Name:  sa["stream_name"].(string),
				Query: sa["query"].(string),
			},
		)
	}

	_, err := client.UpdateDashboard(
		d.Get("project_name").(string),
		d.Get("dashboard_name").(string),
		searchAttributes,
		d.Id(),
	)
	if err != nil {
		log.Println(err)
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
