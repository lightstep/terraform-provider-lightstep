package main

import (
	"github.com/hashicorp/terraform/helper/schema"
)

func resourceStream() *schema.Resource {
  return &schema.Resource{
    Create: resourceStreamCreate,
    Read:   resourceStreamRead,
    Update: resourceStreamUpdate,
    Delete: resourceStreamDelete,

		Schema: map[string]*schema.Schema{
			"organization": &schema.Schema{
				Type:     schema.TypeString,
				Required: true,
			},
			"project": &schema.Schema{
				Type:     schema.TypeString,
				Required: true,
			},
		},
	}
}

func resourceStreamCreate(d *schema.ResourceData, m interface{}) error {
  return resourceStreamRead(d, m)
}

func resourceStreamRead(d *schema.ResourceData, m interface{}) error {
  return nil
}

func resourceStreamUpdate(d *schema.ResourceData, m interface{}) error {
  return resourceStreamRead(d, m)
}

func resourceStreamDelete(d *schema.ResourceData, m interface{}) error {
  return nil
}
