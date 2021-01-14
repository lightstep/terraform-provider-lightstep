package main

import (
	"context"
	"fmt"
	"strings"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/lightstep/terraform-provider-lightstep/lightstep"
)

// these are common across all types of destinations
func resourceDestinationRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	var diags diag.Diagnostics

	client := m.(*lightstep.Client)
	if _, err := client.GetDestination(d.Get("project_name").(string), d.Id()); err != nil {
		return diag.FromErr(fmt.Errorf("Failed to get destination: %v", err))
	}

	return diags
}

func resourceDestinationDelete(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	var diags diag.Diagnostics

	client := m.(*lightstep.Client)
	if err := client.DeleteDestination(d.Get("project_name").(string), d.Id()); err != nil {
		return diag.FromErr(fmt.Errorf("Failed to delete destination: %v", err))
	}

	return diags
}

func splitID(id string) ([]string, error) {
	ids := strings.Split(id, ".")
	if len(ids) != 2 {
		return nil, fmt.Errorf("Error importing lightstep_pagerduty_destination. Expecting an ID formed as '<lightstep_project>.<lightstep_destination_ID>'")
	}
	return ids, nil
}
