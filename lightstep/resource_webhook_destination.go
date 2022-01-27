package lightstep

import (
	"context"
	"fmt"
	"github.com/lightstep/terraform-provider-lightstep/client"
	"strings"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func resourceWebhookDestination() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceWebhookDestinationCreate,
		ReadContext:   resourceDestinationRead,
		DeleteContext: resourceDestinationDelete,
		Importer: &schema.ResourceImporter{
			StateContext: resourceWebhookDestinationImport,
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

func resourceWebhookDestinationCreate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	c := m.(*client.Client)

	dest := client.Destination{
		Type: "destination",
	}
	attrs := client.WebhookAttributes{
		Name:            d.Get("destination_name").(string),
		DestinationType: "webhook",
		URL:             d.Get("url").(string),
	}

	headers, ok := d.GetOk("custom_headers")
	if ok {
		attrs.CustomHeaders = headers.(map[string]interface{})
	}

	dest.Attributes = attrs
	destination, err := c.CreateDestination(ctx, d.Get("project_name").(string), dest)
	if err != nil {
		return diag.FromErr(fmt.Errorf("Failed to create webhook destination %s: %v", destination, err))
	}

	d.SetId(destination.ID)
	return resourceDestinationRead(ctx, d, m)
}

func resourceWebhookDestinationImport(ctx context.Context, d *schema.ResourceData, m interface{}) ([]*schema.ResourceData, error) {
	c := m.(*client.Client)

	ids := strings.Split(d.Id(), ".")
	if len(ids) != 2 {
		return []*schema.ResourceData{}, fmt.Errorf("Error importing lightstep_webhook_destination. Expecting an  ID formed as '<lightstep_project>.<lightstep_destination_ID>'")
	}

	project, id := ids[0], ids[1]
	dest, err := c.GetDestination(ctx, project, id)
	if err != nil {
		return []*schema.ResourceData{}, fmt.Errorf("Failed to get webhook destination: %v", err)
	}

	d.SetId(dest.ID)
	if err := d.Set("project_name", project); err != nil {
		return []*schema.ResourceData{}, fmt.Errorf("Unable to set project_name resource field: %v", err)
	}

	attributes := dest.Attributes.(map[string]interface{})
	if err := d.Set("destination_name", attributes["name"]); err != nil {
		return []*schema.ResourceData{}, fmt.Errorf("Unable to set destination_name resource field: %v", err)
	}

	if err := d.Set("url", attributes["url"]); err != nil {
		return []*schema.ResourceData{}, fmt.Errorf("Unable to set url resource field: %v", err)
	}

	if len(attributes["custom_headers"].(map[string]interface{})) > 0 {
		if err := d.Set("custom_headers", attributes["custom_headers"]); err != nil {
			return []*schema.ResourceData{}, fmt.Errorf("Unable to set custom_headers resource field: %v", err)
		}
	}

	return []*schema.ResourceData{d}, nil
}
