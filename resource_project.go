package main

import (
	"github.com/hashicorp/terraform/helper/schema"
	"github.com/lightstep/terraform-provider-lightstep/lightstep"
)

func resourceProject() *schema.Resource {
	return &schema.Resource{
		Create: resourceProjectCreate,
		Read:   resourceProjectRead,
		Update: resourceProjectUpdate,
		Delete: resourceProjectDelete,
		Exists: resourceProjectExists,

		Schema: map[string]*schema.Schema{
			"project_name": &schema.Schema{
				Type:     schema.TypeString,
				Required: true,
			},
		},
	}
}

func resourceProjectExists(d *schema.ResourceData, m interface{}) (b bool, e error) {
	client := m.(*lightstep.Client)
	_, err := client.ReadProject(
		d.Get("project_name").(string),
	)
	if err != nil {
		return false, err
	}
	return true, nil
}

func resourceProjectCreate(d *schema.ResourceData, m interface{}) error {
	client := m.(*lightstep.Client)
	resp, err := client.CreateProject(
		d.Get("project_name").(string),
	)
	if err != nil {
		return err
	}
	d.SetId(string(resp.Data.ID))
	return resourceProjectRead(d, m)
}

func resourceProjectRead(d *schema.ResourceData, m interface{}) error {
	client := m.(*lightstep.Client)
	_, err := client.ReadProject(
		d.Get("project_name").(string),
	)
	if err != nil {
		return err
	}
	return nil
}

func resourceProjectUpdate(d *schema.ResourceData, m interface{}) error {
	return resourceProjectRead(d, m)
}

func resourceProjectDelete(d *schema.ResourceData, m interface{}) error {
	client := m.(*lightstep.Client)
	err := client.DeleteProject(
		d.Get("project_name").(string),
	)
	if err != nil {
		return err
	}
	d.SetId("")
	return nil
}
