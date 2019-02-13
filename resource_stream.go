package main

import (
  "github.com/hashicorp/terraform/helper/schema"
)

func resourceProject() *schema.Resource {
  return &schema.Resource{
    Create: resourceProjectCreate,
    Read:   resourceProjectRead,
    Update: resourceProjectUpdate,
    Delete: resourceProjectDelete,

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

func resourceProjectCreate(d *schema.ResourceData, m interface{}) error {
  return resourceProjectRead(d, m)
}

func resourceProjectRead(d *schema.ResourceData, m interface{}) error {
  return nil
}

func resourceProjectUpdate(d *schema.ResourceData, m interface{}) error {
  return resourceProjectRead(d, m)
}

func resourceProjectDelete(d *schema.ResourceData, m interface{}) error {
  return nil
}