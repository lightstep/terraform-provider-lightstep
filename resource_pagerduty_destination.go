package main

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/lightstep/terraform-provider-lightstep/lightstep"
)

func resourcePagerdutyDestination() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourcePagerdutyDestinationCreate,
		ReadContext:   resourceDestinationRead,
		DeleteContext: resourceDestinationDelete,
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

func resourcePagerdutyDestinationCreate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
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
		return diag.FromErr(fmt.Errorf("Failed to create pagerduty destination: %v", err))
	}

	d.SetId(destination.ID)
	return resourceDestinationRead(ctx, d, m)
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
		return []*schema.ResourceData{}, fmt.Errorf("Failed to get pagerduty destination: %v", err)
	}

	d.SetId(c.ID)
	if err := d.Set("project_name", project); err != nil {
		return []*schema.ResourceData{}, fmt.Errorf("Unable to set project_name resource field: %v", err)
	}

	attributes := c.Attributes.(map[string]interface{})
	if err := d.Set("destination_name", attributes["name"]); err != nil {
		return []*schema.ResourceData{}, fmt.Errorf("Unable to set destination_name resource field: %v", err)
	}

	if err := d.Set("integration_key", attributes["integration_key"]); err != nil {
		return []*schema.ResourceData{}, fmt.Errorf("Unable to set integration_key resource field: %v", err)
	}

	return []*schema.ResourceData{d}, nil
}
