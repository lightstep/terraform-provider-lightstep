package main

import (
	"fmt"
	"strings"

	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
	"github.com/lightstep/terraform-provider-lightstep/lightstep"
)

func resourceDestination() *schema.Resource {
	return &schema.Resource{
		Create: resourceDestinationCreate,
		Read:   resourceDestinationRead,
		Delete: resourceDestinationDelete,
		Importer: &schema.ResourceImporter{
			State: resourceDestinationImport,
		},
		Schema: map[string]*schema.Schema{
			"project_name": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"destination_type": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"destination_name": {
				Type:     schema.TypeString,
				ForceNew: true,
				Optional: true,
			},
			"url": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"custom_headers": {
				Type:     schema.TypeMap,
				Optional: true,
				ForceNew: true,
			},
		},
	}
}

func resourceDestinationCreate(d *schema.ResourceData, m interface{}) error {
	client := m.(*lightstep.Client)
	attrs, err := setAttributesForDestinationType(d)
	if err != nil {
		return err
	}

	destination, err := client.CreateDestination(
		d.Get("project_name").(string),
		attrs,
	)
	if err != nil {
		return err
	}

	d.SetId(destination.ID)
	return resourceConditionRead(d, m)
}

func setAttributesForDestinationType(d *schema.ResourceData) (interface{}, error) {
	switch d.Get("destination_type").(string) {
	case lightstep.WEBHOOK_DESTINATION_TYPE:
		var attributes lightstep.WebhookAttributes

		name, ok := d.GetOk("destination_name")
		if !ok {
			return nil, fmt.Errorf("Missing required parameter 'name' for %v type\n", lightstep.WEBHOOK_DESTINATION_TYPE)
		}

		url, ok := d.GetOk("url")
		if !ok {
			return nil, fmt.Errorf("Missing required parameter 'url' for %v type\n", lightstep.WEBHOOK_DESTINATION_TYPE)
		}

		customHeaders, ok := d.GetOk("custom_headers")
		if ok {
			attributes.CustomHeaders = customHeaders.(map[string]interface{})
		}
		attributes.Name = name.(string)
		attributes.URL = url.(string)
		attributes.DestinationType = lightstep.WEBHOOK_DESTINATION_TYPE
		return attributes, nil

	default:
		return nil, nil
	}

}

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

func resourceDestinationImport(d *schema.ResourceData, m interface{}) ([]*schema.ResourceData, error) {
	client := m.(*lightstep.Client)

	ids := strings.Split(d.Id(), ".")
	if len(ids) != 2 {
		return []*schema.ResourceData{}, fmt.Errorf("Error importing lightstep_condition. Expecting an  ID formed as '<lightstep_project>.<lightstep_destination_ID>'")
	}

	project, id := ids[0], ids[1]
	c, err := client.GetDestination(project, id)
	if err != nil {
		return []*schema.ResourceData{}, err
	}
	fmt.Printf("project %v, id %v\n", project, id)
	d.SetId(c.ID)
	d.Set("project_name", project)    // nolint  these values are fetched from LS
	d.Set("destination_type", c.Type) // nolint  and known to be valid

	switch c.Type {
	case lightstep.WEBHOOK_DESTINATION_TYPE:
		d.Set("destination_name", c.Attributes.(lightstep.WebhookAttributes).Name) // nolint
		d.Set("url", c.Attributes.(lightstep.WebhookAttributes).URL)               // nolint
		if len(c.Attributes.(lightstep.WebhookAttributes).CustomHeaders) != 0 {
			d.Set("custom_headers", c.Attributes.(lightstep.WebhookAttributes).CustomHeaders) // nolint
		}
	}

	return []*schema.ResourceData{d}, nil
}
