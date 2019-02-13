package main

import (
  "fmt"
	"github.com/hashicorp/terraform/helper/schema"
  "github.com/lightstep/terraform-provider-lightstep/lightstep"
)

func resourceStream() *schema.Resource {
	return &schema.Resource{
		Create: resourceStreamCreate,
		Read:   resourceStreamRead,
		Update: resourceStreamUpdate,
		Delete: resourceStreamDelete,
    // Exists: resourceStreamExists,

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

// func resourceStreamExists(d *schema.ResourceData, m interface{}) (b bool, e error) {
//   return false, nil
// }

func resourceStreamCreate(d *schema.ResourceData, m interface{}) error {
  client := m.(*lightstep.Client)

  client.CreateSearch(
    d.Get("project"),
    d.Get("name"),
    d.Get("query"),
  )
  fmt.Println(client)
  //customData map[string]interface{},
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
