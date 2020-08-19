package main

import (
	"fmt"
	"strings"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
	"github.com/lightstep/terraform-provider-lightstep/lightstep"
)

func resourceStream() *schema.Resource {
	return &schema.Resource{
		Create: resourceStreamCreate,
		Read:   resourceStreamRead,
		Update: resourceStreamUpdate,
		Delete: resourceStreamDelete,
		Importer: &schema.ResourceImporter{
			State: resourceStreamImport,
		},
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
				ForceNew: true,
			},
			"custom_data": {
				Type:     schema.TypeMap,
				Optional: true,
			},
		},
		Timeouts: &schema.ResourceTimeout{
			Create: schema.DefaultTimeout(5 * time.Second),
		},
	}
}

func resourceStreamCreate(d *schema.ResourceData, m interface{}) error {
	client := m.(*lightstep.Client)

	return resource.Retry(d.Timeout(schema.TimeoutCreate), func() *resource.RetryError {
		stream, err := client.CreateStream(
			d.Get("project_name").(string),
			d.Get("stream_name").(string),
			d.Get("query").(string),
			d.Get("custom_data").(map[string]interface{}),
		)
		if err != nil {
			// Fix until lock error is resolved
			if strings.Contains(err.Error(), "Internal Server Error") {
				return resource.RetryableError(fmt.Errorf("Expected Creation of stream but not done yet"))
			} else {
				return resource.NonRetryableError(fmt.Errorf("Error creating stream: %s", err))
			}
		}

		d.SetId(stream.ID)
		return resource.NonRetryableError(resourceStreamRead(d, m))
	})
}

func resourceStreamRead(d *schema.ResourceData, m interface{}) error {
	client := m.(*lightstep.Client)
	_, err := client.GetStream(
		d.Get("project_name").(string),
		d.Id(),
	)
	if err != nil {
		return err
	}
	return nil
}

func resourceStreamUpdate(d *schema.ResourceData, m interface{}) error {
	client := m.(*lightstep.Client)
	if d.HasChange("query") {
		if err := client.DeleteStream(
			d.Get("project_name").(string),
			d.Id(),
		); err != nil {
			return err
		}
	}
	return resourceStreamCreate(d, m)
}

func resourceStreamDelete(d *schema.ResourceData, m interface{}) error {
	client := m.(*lightstep.Client)
	err := client.DeleteStream(
		d.Get("project_name").(string),
		d.Id(),
	)
	if err != nil {
		return err
	}
	d.SetId("")
	return nil
}

func resourceStreamImport(d *schema.ResourceData, m interface{}) ([]*schema.ResourceData, error) {
	client := m.(*lightstep.Client)

	ids := strings.Split(d.Id(), ".")
	if len(ids) != 2 {
		return []*schema.ResourceData{}, fmt.Errorf("Error importing lightstep_stream. Expecting an  ID formed as '<lightstep_project>.<stream_id>'. Got: %v", d.Id())
	}

	project, id := ids[0], ids[1]
	stream, err := client.GetStream(project, id)
	if err != nil {
		return []*schema.ResourceData{}, err
	}

	d.SetId(id)
	d.Set("project_name", project)               //nolint project_name is already valid since it is used in API call above
	d.Set("stream_name", stream.Attributes.Name) //nolint stream_name or query because they are received from API call
	d.Set("query", stream.Attributes.Query)      //nolint

	return []*schema.ResourceData{d}, nil
}
