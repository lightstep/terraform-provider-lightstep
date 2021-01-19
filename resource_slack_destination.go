package main

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/lightstep/terraform-provider-lightstep/lightstep"
)

func resourceSlackDestination() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceSlackDestinationCreate,
		ReadContext:   resourceSlackDestinationRead,
		DeleteContext: resourceDestinationDelete,
		Importer: &schema.ResourceImporter{
			State: resourceSlackDestinationImport,
		},
		Schema: map[string]*schema.Schema{
			"project_name": {
				Type:        schema.TypeString,
				Required:    true,
				ForceNew:    true,
				Description: "Lightstep project name",
			},
			"channel": {
				Type:        schema.TypeString,
				Required:    true,
				ForceNew:    true,
				Description: "One of: slack channel name (#channel), channel ID, handle name (@user).",
			},
		},
	}
}

func resourceSlackDestinationRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	var diags diag.Diagnostics

	client := m.(*lightstep.Client)
	dest, err := client.GetDestination(d.Get("project_name").(string), d.Id())
	if err != nil {
		d.SetId("")
		return diags
	}

	if err := d.Set("channel", dest.Attributes.(map[string]interface{})["channel"]); err != nil {
		return diag.FromErr(fmt.Errorf("Unable to set channel resource field: %v", err))
	}

	return diags
}

func resourceSlackDestinationCreate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	client := m.(*lightstep.Client)
	attrs := lightstep.SlackAttributes{
		Channel:         d.Get("channel").(string),
		DestinationType: "slack",
	}
	dest := lightstep.Destination{
		Type:       "destination",
		Attributes: attrs,
	}

	destination, err := client.CreateDestination(d.Get("project_name").(string), dest)
	if err != nil {
		return diag.FromErr(fmt.Errorf("Failed to create slack destination %v: %v", attrs.Channel, err))
	}

	d.SetId(destination.ID)
	return resourceDestinationRead(ctx, d, m)
}

func resourceSlackDestinationImport(d *schema.ResourceData, m interface{}) ([]*schema.ResourceData, error) {
	client := m.(*lightstep.Client)

	ids, err := splitID(d.Id())
	if err != nil {
		return nil, err
	}

	project, id := ids[0], ids[1]
	c, err := client.GetDestination(project, id)
	if err != nil {
		return []*schema.ResourceData{}, fmt.Errorf("Failed to get slack destination: %v", err)
	}

	d.SetId(c.ID)
	if err := d.Set("project_name", project); err != nil {
		return []*schema.ResourceData{}, fmt.Errorf("Unable to set project_name resource field: %v", err)
	}

	attributes := c.Attributes.(map[string]interface{})
	if err := d.Set("channel", attributes["channel"]); err != nil {
		return []*schema.ResourceData{}, fmt.Errorf("Unable to set channel resource field: %v", err)
	}

	return []*schema.ResourceData{d}, nil
}
