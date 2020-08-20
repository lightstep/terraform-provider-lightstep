package main

import (
	"fmt"
	"strings"

	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
	"github.com/lightstep/terraform-provider-lightstep/lightstep"
)

func resourceSlackDestination() *schema.Resource {
	return &schema.Resource{
		Create: resourceSlackDestinationCreate,
		Read:   resourceDestinationRead,
		Delete: resourceDestinationDelete,
		Importer: &schema.ResourceImporter{
			State: resourceSlackDestinationImport,
		},
		Schema: map[string]*schema.Schema{
			"project_name": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"channel": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
		},
	}
}

func resourceSlackDestinationCreate(d *schema.ResourceData, m interface{}) error {
	client := m.(*lightstep.Client)

	dest := lightstep.Destination{
		Type: "destination",
	}

	attrs := lightstep.SlackAttributes{
		Channel:         d.Get("channel").(string),
		DestinationType: "slack",
	}

	dest.Attributes = attrs

	destination, err := client.CreateDestination(d.Get("project_name").(string), dest)
	if err != nil {
		return fmt.Errorf("Error creating destination to %v.\nErr:%v\n", attrs.Channel, err)
	}

	d.SetId(destination.ID)
	return resourceDestinationRead(d, m)
}

func resourceSlackDestinationImport(d *schema.ResourceData, m interface{}) ([]*schema.ResourceData, error) {
	client := m.(*lightstep.Client)

	ids := strings.Split(d.Id(), ".")
	if len(ids) != 2 {
		return []*schema.ResourceData{}, fmt.Errorf("Error importing lightstep_slack_destination. Expecting an  ID formed as '<lightstep_project>.<lightstep_destination_ID>'")
	}

	project, id := ids[0], ids[1]
	c, err := client.GetDestination(project, id)
	if err != nil {
		return []*schema.ResourceData{}, err
	}
	d.SetId(c.ID)

	attributes := c.Attributes.(map[string]interface{})

	d.Set("project_name", project)          // nolint  these values are fetched from LS
	d.Set("channel", attributes["channel"]) // nolint  and known to be valid

	return []*schema.ResourceData{d}, nil
}
