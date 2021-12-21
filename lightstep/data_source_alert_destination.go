package lightstep

import (
	"context"
	"fmt"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/lightstep/terraform-provider-lightstep/client"
	"net/http"
)

func dataSourceAlertDestination() *schema.Resource {
	return &schema.Resource{
		Description: "Use this data source to retrieve information about an existing stream for use in other resources.",
		ReadContext: dataSourceAlertDestinationRead,
		Schema: map[string]*schema.Schema{
			"project_name": {
				Type:     schema.TypeString,
				Required: true,
			},
			"alert_destination_id": {
				Type:     schema.TypeString,
				Required: true,
			},
		},
	}
}

func dataSourceAlertDestinationRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	var diags diag.Diagnostics

	c := m.(*client.Client)
	dest, err := c.GetDestination(d.Get("project_name").(string), d.Get("alert_destination_id").(string))
	if err != nil {
		apiErr := err.(client.APIResponseCarrier)
		if apiErr.GetHTTPResponse().StatusCode == http.StatusNotFound {
			d.SetId("")
			return diags
		}
		return diag.FromErr(fmt.Errorf("Failed to get destination: %v\n", apiErr))
	}
	d.SetId(dest.ID)
	return diags
}
