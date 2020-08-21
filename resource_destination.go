package main

import (
	"fmt"
	"strings"

	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
	"github.com/lightstep/terraform-provider-lightstep/lightstep"
)

// these are common across all types of destinations
func resourceDestinationRead(d *schema.ResourceData, m interface{}) error {
	client := m.(*lightstep.Client)
	_, err := client.GetDestination(d.Get("project_name").(string), d.Id())
	return err
}

func resourceDestinationDelete(d *schema.ResourceData, m interface{}) error {
	client := m.(*lightstep.Client)
	err := client.DeleteDestination(d.Get("project_name").(string), d.Id())
	return err
}

func splitID(id string) ([]string, error) {
	ids := strings.Split(id, ".")
	if len(ids) != 2 {
		return nil, fmt.Errorf("Error importing lightstep_pagerduty_destination. Expecting an ID formed as '<lightstep_project>.<lightstep_destination_ID>'")
	}
	return ids, nil
}
