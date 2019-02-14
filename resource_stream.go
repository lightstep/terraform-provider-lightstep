package main

import (
	"github.com/hashicorp/terraform/helper/schema"
	"github.com/lightstep/terraform-provider-lightstep/lightstep"
	"log"
)

func resourceStream() *schema.Resource {
	return &schema.Resource{
		Create: resourceStreamCreate,
		Read:   resourceStreamRead,
		Update: resourceStreamUpdate,
		Delete: resourceStreamDelete,
		Exists: resourceStreamExists,

		Schema: map[string]*schema.Schema{
			"project_name": &schema.Schema{
				Type:     schema.TypeString,
				Required: true,
			},
			"stream_name": &schema.Schema{
				Type:     schema.TypeString,
				Required: true,
			},
			"query": &schema.Schema{
				Type:     schema.TypeString,
				Required: true,
			},
      "custom_data": &schema.Schema{
        Type:     schema.TypeMap,
        Optional: true,
      },
		},
	}
}

func resourceStreamExists(d *schema.ResourceData, m interface{}) (b bool, e error) {
	client := m.(*lightstep.Client)

	if _, err := client.GetSearch(
		d.Get("project_name").(string),
		d.Id(),
	); err != nil {
		return false, err
	}

	return true, nil
}

func resourceStreamCreate(d *schema.ResourceData, m interface{}) error {
	client := m.(*lightstep.Client)
	resp, err := client.CreateSearch(
		d.Get("project_name").(string),
		d.Get("stream_name").(string),
		d.Get("query").(string),
		d.Get("custom_data").(map[string]interface{}),
	)
	if err != nil {
		log.Println(err)
		return err
	}
	d.SetId(string(resp.Data.ID))
	return resourceStreamRead(d, m)
}

func resourceStreamRead(d *schema.ResourceData, m interface{}) error {
	client := m.(*lightstep.Client)
	_, err := client.GetSearch(
		d.Get("project_name").(string),
		d.Id(),
	)
	if err != nil {
		return err
	}
	return nil
}

func resourceStreamUpdate(d *schema.ResourceData, m interface{}) error {
	return resourceStreamCreate(d, m)
}

func resourceStreamDelete(d *schema.ResourceData, m interface{}) error {

	client := m.(*lightstep.Client)
	err := client.DeleteSearch(
		d.Get("project_name").(string),
		d.Id(),
	)
	if err != nil {
		return err
	}
	d.SetId("")
	return nil
}
