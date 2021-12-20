package lightstep

import (
	"context"
	"fmt"
	"github.com/lightstep/terraform-provider-lightstep/client"
	"net/http"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
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

func resourceSlackDestinationRead(_ context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	var diags diag.Diagnostics

	c := m.(*client.Client)
	dest, err := c.GetDestination(d.Get("project_name").(string), d.Id())
	if err != nil {
		apiErr := err.(client.APIResponseCarrier)
		if apiErr.GetHTTPResponse().StatusCode == http.StatusNotFound {
			d.SetId("")
			return diags
		}
		return diag.FromErr(fmt.Errorf("Failed to get slack destination: %v\n", apiErr))
	}

	if err := d.Set("channel", dest.Attributes.(map[string]interface{})["channel"]); err != nil {
		return diag.FromErr(fmt.Errorf("Unable to set channel resource field: %v", err))
	}

	return diags
}

func resourceSlackDestinationCreate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	c := m.(*client.Client)
	attrs := client.SlackAttributes{
		Channel:         d.Get("channel").(string),
		DestinationType: "slack",
	}
	dest := client.Destination{
		Type:       "destination",
		Attributes: attrs,
	}

	destination, err := c.CreateDestination(d.Get("project_name").(string), dest)
	if err != nil {
		return diag.FromErr(fmt.Errorf("Failed to create slack destination %v: %v", attrs.Channel, err))
	}

	d.SetId(destination.ID)
	return resourceDestinationRead(ctx, d, m)
}

func resourceSlackDestinationImport(d *schema.ResourceData, m interface{}) ([]*schema.ResourceData, error) {
	c := m.(*client.Client)

	ids, err := splitID(d.Id())
	if err != nil {
		return nil, err
	}

	project, id := ids[0], ids[1]
	dest, err := c.GetDestination(project, id)
	if err != nil {
		return []*schema.ResourceData{}, fmt.Errorf("Failed to get slack destination: %v", err)
	}

	d.SetId(dest.ID)
	if err := d.Set("project_name", project); err != nil {
		return []*schema.ResourceData{}, fmt.Errorf("Unable to set project_name resource field: %v", err)
	}

	attributes := dest.Attributes.(map[string]interface{})
	if err := d.Set("channel", attributes["channel"]); err != nil {
		return []*schema.ResourceData{}, fmt.Errorf("Unable to set channel resource field: %v", err)
	}

	return []*schema.ResourceData{d}, nil
}
