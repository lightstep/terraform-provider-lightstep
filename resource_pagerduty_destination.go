package main

import (
	"fmt"
	"strings"

	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
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
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"destination_name": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"integration_key": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
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

	ids := strings.Split(d.Id(), ".")
	if len(ids) != 2 {
		return []*schema.ResourceData{}, fmt.Errorf("Error importing lightstep_pagerduty_destination. Expecting an  ID formed as '<lightstep_project>.<lightstep_destination_ID>'")
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
