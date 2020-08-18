package main

import (
	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
	"github.com/lightstep/terraform-provider-lightstep/lightstep"
)

// these are common across all types of destinations
func resourceDestinationRead(d *schema.ResourceData, m interface{}) error {
	client := m.(*lightstep.Client)
	_, err := client.GetDestination(d.Get("project_name").(string), d.Id())
	return err
}

func resourceDestinationDelete(d *schema.ResourceData, m interface{}) error {
	client := m.(*lightstep.Client)
	err := client.DeleteDestination(d.Get("project_name").(string), d.Id())
	return err
}
