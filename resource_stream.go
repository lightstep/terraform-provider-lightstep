package main

import (
	"github.com/hashicorp/terraform/helper/schema"
  "github.com/lightstep/terraform-provider-lightstep/lightstep"
)

func resourceStream() *schema.Resource {
	return &schema.Resource{
		Create: resourceStreamCreate,
		Read:   resourceStreamRead,
		Update: resourceStreamUpdate,
		Delete: resourceStreamDelete,
    Exists: resourceStreamExists,

		Schema: map[string]*schema.Schema{
			"project": &schema.Schema{
				Type:     schema.TypeString,
				Required: true,
			},
      "stream_id": &schema.Schema{
        Type:     schema.TypeString,
        Optional: true,
      },
      "name": &schema.Schema{
        Type:     schema.TypeString,
        Required: true,
      },
      "query": &schema.Schema{
        Type:     schema.TypeString,
        Required: true,
      },
		},
	}
}

func resourceStreamExists(d *schema.ResourceData, m interface{}) (b bool, e error) {
  return false, nil
}

func resourceStreamCreate(d *schema.ResourceData, m interface{}) error {
  client := m.(*lightstep.Client)
  _, err := client.CreateSearch(
    d.Get("project").(string),
    d.Get("name").(string),
    d.Get("query").(string),
    nil,
  )
  if err != nil {
    return err
  }
	return resourceStreamRead(d, m)
}

func resourceStreamRead(d *schema.ResourceData, m interface{}) error {
	client := m.(*lightstep.Client)
  _, err := client.GetSearch(
    d.Get("project").(string),
    d.Get("stream_id").(string),
  )
  if err != nil {
    return err
  }
  return nil
}

func resourceStreamUpdate(d *schema.ResourceData, m interface{}) error {
	return resourceStreamRead(d, m)
}

func resourceStreamDelete(d *schema.ResourceData, m interface{}) error {
	return nil
}
