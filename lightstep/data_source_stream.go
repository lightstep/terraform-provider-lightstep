package lightstep

import (
	"context"
	"fmt"
	"net/http"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/lightstep/terraform-provider-lightstep/client"
)

func dataSourceStream() *schema.Resource {
	return &schema.Resource{
		Description: "Use this data source to retrieve information about an existing stream for use in other resources.",
		ReadContext: dataSourceLightstepStreamRead,
		Schema: map[string]*schema.Schema{
			"project_name": {
				Type:     schema.TypeString,
				Required: true,
			},
			"stream_id": {
				Type:     schema.TypeString,
				Required: true,
			},
			// Computed
			"stream_name": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"stream_query": {
				Description: "Stream query",
				Type:        schema.TypeString,
				Computed:    true,
			},
		},
	}
}

func dataSourceLightstepStreamRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	c := m.(*client.Client)
	s, err := c.GetStream(ctx, d.Get("project_name").(string), d.Get("stream_id").(string))
	if err != nil {
		apiErr := err.(client.APIResponseCarrier)
		if apiErr.GetHTTPResponse().StatusCode == http.StatusNotFound {
			d.SetId("")
			return diag.FromErr(fmt.Errorf("Stream not found: %v\n", apiErr))
		}
		return diag.FromErr(fmt.Errorf("Failed to get stream: %v\n", apiErr))
	}
	d.SetId(s.ID)
	if err := d.Set("stream_name", s.Attributes.Name); err != nil {
		return diag.FromErr(err)
	}

	if err := d.Set("stream_query", s.Attributes.Query); err != nil {
		return diag.FromErr(err)
	}
	return nil
}
