package lightstep

import (
	"context"
	"fmt"
	"net/http"
	"strings"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/lightstep/terraform-provider-lightstep/client"
)

func errorIsNotFound(err error) bool {
	if err == nil {
		return false
	}
	apiErr, ok := err.(client.APIResponseCarrier)
	if !ok {
		return false
	}
	resp := apiErr.GetHTTPResponse()
	if resp == nil {
		return false
	}
	return resp.StatusCode == http.StatusNotFound
}

// these are common across all types of destinations
func resourceDestinationRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	c := m.(*client.Client)
	dest, err := c.GetDestination(ctx, d.Get("project_name").(string), d.Id())
	if err != nil {
		if errorIsNotFound(err) {
			d.SetId("")
			return diag.Diagnostics{}
		}
		return diag.FromErr(fmt.Errorf("failed to get destination: %v", err))
	}
	d.SetId(dest.ID)
	return diag.Diagnostics{}
}

func resourceDestinationDelete(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	var diags diag.Diagnostics

	client := m.(*client.Client)
	if err := client.DeleteDestination(ctx, d.Get("project_name").(string), d.Id()); err != nil {
		return diag.FromErr(fmt.Errorf("failed to delete destination: %v", err))
	}

	return diags
}

func splitID(id string) ([]string, error) {
	ids := strings.Split(id, ".")
	if len(ids) != 2 {
		return nil, fmt.Errorf("error importing lightstep_pagerduty_destination. Expecting an ID formed as '<lightstep_project>.<lightstep_destination_ID>'")
	}
	return ids, nil
}
