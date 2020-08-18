package main

import (
	"fmt"
	"strings"

	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
	"github.com/lightstep/terraform-provider-lightstep/lightstep"
)

func resourceDestination() *schema.Resource {
	return &schema.Resource{
		Create: resourceDestinationCreate,
		Read:   resourceDestinationRead,
		Delete: resourceDestinationDelete,
		Importer: &schema.ResourceImporter{
			State: resourceDestinationImport,
		},
		Schema: map[string]*schema.Schema{
			"project_name": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"destination_type": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"destination_name": {
				Type:     schema.TypeString,
				Optional: true,
				ForceNew: true,
			},
			"url": {
				Type:     schema.TypeString,
				Optional: true,
				ForceNew: true,
			},
		},
	}
}

func resourceDestinationCreate(d *schema.ResourceData, m interface{}) error {
	client := m.(*lightstep.Client)
	err := validateAttributesForType(d)
	if err != nil {
		return err
	}

	attributes, err := getDestinationAttributesForType(d)
	if err != nil {
		return err
	}

	destination, err := client.CreateDestination(
		d.Get("project_name").(string),
		d.Get("destination_type").(string),
		attributes)
	if err != nil {
		return fmt.Errorf("Error creating %v.\nErr:%v\n", destination, err)
	}

	d.SetId(destination.ID)
	return resourceDestinationRead(d, m)
}

func getRequiredAttributesForType(destType string) []string {
	destinationTypeToAttributes := map[string][]string{
		"webhook": {
			"destination_name",
			"url",
		},
		"pagerduty": {
			"integration_key",
			"destination_name",
		},
		"slack": {
			"channel",
			"scope",
		},
	}
	return destinationTypeToAttributes[destType]

}

// performing validation of attributes for type here since
// terraform's ValidateFunc does not allow you to inspect other fields
func validateAttributesForType(d *schema.ResourceData) error {
	requiredAttributes := map[string][]string{
		"webhook": {
			"destination_name",
			"url",
		},
		"pagerduty": {
			"integration_key",
			"destination_name",
		},
		"slack": {
			"channel",
			"scope",
		},
	}

	destinationType := d.Get("destination_type").(string)
	requiredAttrForType := requiredAttributes[destinationType]
	for _, attr := range requiredAttrForType {
		_, ok := d.GetOk(attr)
		if !ok {
			return fmt.Errorf("Missing required attribute %v for destination type %v", attr, destinationType)
		}
	}
	return nil
}

func getDestinationAttributesForType(d *schema.ResourceData) (map[string]interface{}, error) {
	destinationType := d.Get("destination_type").(string)

	attributes := map[string]interface{}{}

	requiredAttrs := getRequiredAttributesForType(destinationType)
	for _, attr := range requiredAttrs {
		v, ok := d.GetOk(attr)
		if !ok {
			return nil, fmt.Errorf("Missing required parameter %v. Got: %v\n", attr, v)
		}
		attributes[attr] = v
	}

	return attributes, nil

}

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

func resourceDestinationImport(d *schema.ResourceData, m interface{}) ([]*schema.ResourceData, error) {
	client := m.(*lightstep.Client)

	ids := strings.Split(d.Id(), ".")
	if len(ids) != 2 {
		return []*schema.ResourceData{}, fmt.Errorf("Error importing lightstep_condition. Expecting an  ID formed as '<lightstep_project>.<lightstep_destination_ID>'")
	}

	project, id := ids[0], ids[1]
	c, err := client.GetDestination(project, id)
	if err != nil {
		return []*schema.ResourceData{}, err
	}
	d.SetId(c.ID)
	d.Set("project_name", project)    // nolint  these values are fetched from LS
	d.Set("destination_type", c.Type) // nolint  and known to be valid
	destinationType := d.Get("destination_type").(string)
	requiredAttrs := getRequiredAttributesForType(destinationType)

	terraformAttributes := map[string]string{}
	var attributes = map[string]interface{}{}
	for _, attr := range requiredAttrs {
		v, ok := terraformAttributes[attr]
		if !ok {
			return []*schema.ResourceData{}, fmt.Errorf("Missing required parameter %v. Got: %v\n", attr, v)
		}
		attributes[attr] = v
	}

	d.Set("attributes", attributes) //no lint

	return []*schema.ResourceData{d}, nil
}
