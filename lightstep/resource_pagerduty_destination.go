package lightstep

import (
	"context"
	"fmt"
	"github.com/lightstep/terraform-provider-lightstep/client"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func resourcePagerdutyDestination() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourcePagerdutyDestinationCreate,
		ReadContext:   resourceDestinationRead,
		DeleteContext: resourceDestinationDelete,
		Importer: &schema.ResourceImporter{
			StateContext: resourcePagerdutyDestinationImport,
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
	c := m.(*client.Client)
	destination, err := c.CreateDestination(ctx, d.Get("project_name").(string),
		client.Destination{
			Type: "destination",
			Attributes: client.PagerdutyAttributes{
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

func resourcePagerdutyDestinationImport(ctx context.Context, d *schema.ResourceData, m interface{}) ([]*schema.ResourceData, error) {
	c := m.(*client.Client)

	ids, err := splitID(d.Id())
	if err != nil {
		return nil, err
	}

	project, id := ids[0], ids[1]
	dest, err := c.GetDestination(ctx, project, id)
	if err != nil {
		return []*schema.ResourceData{}, fmt.Errorf("Failed to get pagerduty destination: %v", err)
	}

	d.SetId(dest.ID)
	if err := d.Set("project_name", project); err != nil {
		return []*schema.ResourceData{}, fmt.Errorf("Unable to set project_name resource field: %v", err)
	}

	attributes := dest.Attributes.(map[string]interface{})
	if err := d.Set("destination_name", attributes["name"]); err != nil {
		return []*schema.ResourceData{}, fmt.Errorf("Unable to set destination_name resource field: %v", err)
	}

	if err := d.Set("integration_key", attributes["integration_key"]); err != nil {
		return []*schema.ResourceData{}, fmt.Errorf("Unable to set integration_key resource field: %v", err)
	}

	return []*schema.ResourceData{d}, nil
}
