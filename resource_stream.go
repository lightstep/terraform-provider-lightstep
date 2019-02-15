package main

import (
	"github.com/hashicorp/terraform/helper/schema"
	"github.com/hashicorp/terraform/helper/resource"
	"github.com/lightstep/terraform-provider-lightstep/lightstep"
	"fmt"
	"time"
	"strings"
)

func resourceStream() *schema.Resource {
	return &schema.Resource{
		Create: resourceStreamCreate,
		Read:   resourceStreamRead,
		Update: resourceStreamUpdate,
		Delete: resourceStreamDelete,
		Exists: resourceStreamExists,

		Schema: map[string]*schema.Schema{
			"project_name": {
				Type:     schema.TypeString,
				Required: true,
			},
			"stream_name": {
				Type:     schema.TypeString,
				Required: true,
			},
			"query": {
				Type:     schema.TypeString,
				Required: true,
			},
			"custom_data": {
				Type:     schema.TypeMap,
				Optional: true,
			},
		},
		Timeouts: &schema.ResourceTimeout{
       Create: schema.DefaultTimeout(2 * time.Second),
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
	return resource.Retry(d.Timeout(schema.TimeoutCreate), func() *resource.RetryError {
      resp, err := client.CreateSearch(
				d.Get("project_name").(string),
				d.Get("stream_name").(string),
				d.Get("query").(string),
				d.Get("custom_data").(map[string]interface{}),
			)
			if err != nil {
				// Fix until lock error is resolved
				if strings.Contains(err.Error(),"Internal Server Error") {
					return resource.RetryableError(fmt.Errorf("Expected Creation of stream but not done yet"))
				} else {
					return resource.NonRetryableError(fmt.Errorf("Error creating stream: %s", err))
				}
			}
      d.SetId(string(resp.Data.ID))
      return resource.NonRetryableError(resourceStreamRead(d, m))
  })
}

func resourceStreamRead(d *schema.ResourceData, m interface{}) error {
	client := m.(*lightstep.Client)
	_, err := client.GetSearch(
		d.Get("project_name").(string),
		d.Id(),
	)
	if err != nil {

		return fmt.Errorf("Error reading stream: %s", err)
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
