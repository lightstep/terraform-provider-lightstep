package main

import (
  "log"
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
      "project": &schema.Schema{
        Type:     schema.TypeString,
        Required: true,
      },
    },
  }
}

func resourceProjectExists(d *schema.ResourceData, m interface{}) (b bool, e error) {
  return false, nil
}

func resourceProjectCreate(d *schema.ResourceData, m interface{}) error {
  client := m.(*lightstep.Client)
  _, err := client.CreateProject(
    d.Get("project").(string),
  )
  if err != nil {
    return err
  }
  return resourceProjectRead(d, m)
}

func resourceProjectRead(d *schema.ResourceData, m interface{}) error {
  client := m.(*lightstep.Client)
  log.Println(client)
  return nil
}

func resourceProjectUpdate(d *schema.ResourceData, m interface{}) error {
  return resourceProjectRead(d, m)
}

func resourceProjectDelete(d *schema.ResourceData, m interface{}) error {
  return nil
}
