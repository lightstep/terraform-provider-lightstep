package main

import (
	"fmt"
	"strings"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/lightstep/terraform-provider-lightstep/lightstep"
)

func resourceWebhookDestination() *schema.Resource {
	return &schema.Resource{
		Create: resourceWebhookDestinationCreate,
		Read:   resourceDestinationRead,
		Delete: resourceDestinationDelete,
		Importer: &schema.ResourceImporter{
			State: resourceWebhookDestinationImport,
		},
		Schema: map[string]*schema.Schema{
			"project_name": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"destination_name": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
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

func resourceWebhookDestinationCreate(d *schema.ResourceData, m interface{}) error {
	client := m.(*lightstep.Client)

	dest := lightstep.Destination{
		Type: "destination",
	}

	attrs := lightstep.WebhookAttributes{
		Name:            d.Get("destination_name").(string),
		DestinationType: "webhook",
		URL:             d.Get("url").(string),
	}

	headers, ok := d.GetOk("custom_headers")
	if ok {
		attrs.CustomHeaders = headers.(map[string]interface{})
	}

	dest.Attributes = attrs

	destination, err := client.CreateDestination(d.Get("project_name").(string), dest)
	if err != nil {
		return fmt.Errorf("Error creating %v.\nErr:%v\n", destination, err)
	}

	d.SetId(destination.ID)
	return resourceDestinationRead(d, m)
}

func resourceWebhookDestinationImport(d *schema.ResourceData, m interface{}) ([]*schema.ResourceData, error) {
	client := m.(*lightstep.Client)

	ids := strings.Split(d.Id(), ".")
	if len(ids) != 2 {
		return []*schema.ResourceData{}, fmt.Errorf("Error importing lightstep_webhook_destination. Expecting an  ID formed as '<lightstep_project>.<lightstep_destination_ID>'")
	}

	project, id := ids[0], ids[1]
	c, err := client.GetDestination(project, id)
	if err != nil {
		return []*schema.ResourceData{}, err
	}
	d.SetId(c.ID)

	attributes := c.Attributes.(map[string]interface{})

	d.Set("project_name", project)                // nolint  these values are fetched from LS
	d.Set("destination_name", attributes["name"]) // nolint  and known to be valid
	d.Set("url", attributes["url"])               // nolint

	if len(attributes["custom_headers"].(map[string]interface{})) > 0 {
		d.Set("custom_headers", attributes["custom_headers"]) // nolint
	}

	return []*schema.ResourceData{d}, nil
}
