package main

import (
	"fmt"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/lightstep/terraform-provider-lightstep/lightstep"
)

func resourcePagerdutyDestination() *schema.Resource {
	return &schema.Resource{
		Create: resourcePagerdutyDestinationCreate,
		Read:   resourceDestinationRead,
		Delete: resourceDestinationDelete,
		Importer: &schema.ResourceImporter{
			State: resourcePagerdutyDestinationImport,
		},
		Schema: map[string]*schema.Schema{
			"project_name": {
				Type:        schema.TypeString,
				Required:    true,
				ForceNew:    true,
				Description: "Lightstep project name",
			},
			"destination_name": {
				Type:        schema.TypeString,
				Required:    true,
				ForceNew:    true,
				Description: "Lightstep destination name",
			},
			"integration_key": {
				Type:        schema.TypeString,
				Required:    true,
				ForceNew:    true,
				Description: "PagerDuty Service Integration Key. To create one follow the docs here - https://support.pagerduty.com/docs/services-and-integrations#add-integrations-to-an-existing-service",
			},
		},
	}
}

func resourcePagerdutyDestinationCreate(d *schema.ResourceData, m interface{}) error {
	client := m.(*lightstep.Client)

	destination, err := client.CreateDestination(d.Get("project_name").(string),
		lightstep.Destination{
			Type: "destination",
			Attributes: lightstep.PagerdutyAttributes{
				Name:            d.Get("destination_name").(string),
				IntegrationKey:  d.Get("integration_key").(string),
				DestinationType: "pagerduty",
			},
		})
	if err != nil {
		return fmt.Errorf("Error creating %v.\nErr:%v\n", destination, err)
	}

	d.SetId(destination.ID)
	return resourceDestinationRead(d, m)
}

func resourcePagerdutyDestinationImport(d *schema.ResourceData, m interface{}) ([]*schema.ResourceData, error) {
	client := m.(*lightstep.Client)

	ids, err := splitID(d.Id())
	if err != nil {
		return nil, err
	}

	project, id := ids[0], ids[1]
	c, err := client.GetDestination(project, id)
	if err != nil {
		return []*schema.ResourceData{}, err
	}
	d.SetId(c.ID)

	attributes := c.Attributes.(map[string]interface{})

	d.Set("project_name", project)                          // nolint  these values are fetched from LS
	d.Set("destination_name", attributes["name"])           // nolint  and known to be valid
	d.Set("integration_key", attributes["integration_key"]) // nolint

	return []*schema.ResourceData{d}, nil
}
