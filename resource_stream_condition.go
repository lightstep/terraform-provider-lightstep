package main

import (
	"fmt"
	"strings"

	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
	"github.com/lightstep/terraform-provider-lightstep/lightstep"
)

func resourceStreamCondition() *schema.Resource {
	return &schema.Resource{
		Create: resourceStreamConditionCreate,
		Read:   resourceStreamConditionRead,
		Delete: resourceStreamConditionDelete,
		Update: resourceStreamConditionUpdate,
		Importer: &schema.ResourceImporter{
			State: resourceStreamConditionImport,
		},
		Schema: map[string]*schema.Schema{
			"project_name": {
				Type:     schema.TypeString,
				Required: true,
			},
			"condition_name": {
				Type:     schema.TypeString,
				Required: true,
			},
			"expression": {
				Type:     schema.TypeString,
				Required: true,
			},
			"stream_id": {
				Type:     schema.TypeString,
				Required: true,
			},
			"evaluation_window_ms": {
				Type:     schema.TypeInt,
				Required: true,
			},
		},
	}
}

func resourceStreamConditionCreate(d *schema.ResourceData, m interface{}) error {
	client := m.(*lightstep.Client)

	condition, err := client.CreateStreamCondition(
		d.Get("project_name").(string),
		d.Get("condition_name").(string),
		d.Get("expression").(string),
		d.Get("evaluation_window_ms").(int),
		d.Get("stream_id").(string))

	if err != nil {
		return err
	}
	d.SetId(condition.ID)
	return resourceStreamConditionRead(d, m)
}

func resourceStreamConditionRead(d *schema.ResourceData, m interface{}) error {
	client := m.(*lightstep.Client)
	_, err := client.GetStreamCondition(d.Get("project_name").(string), d.Id())
	if err != nil {
		return err
	}
	return nil
}

func resourceStreamConditionDelete(d *schema.ResourceData, m interface{}) error {
	client := m.(*lightstep.Client)
	err := client.DeleteStreamCondition(d.Get("project_name").(string), d.Id())
	return err
}

func resourceStreamConditionUpdate(d *schema.ResourceData, m interface{}) error {
	client := m.(*lightstep.Client)

	attrs := lightstep.StreamConditionAttributes{
		Name:               d.Get("condition_name").(string),
		EvaluationWindowMS: d.Get("evaluation_window_ms").(int),
		Expression:         d.Get("expression").(string),
	}

	_, err := client.UpdateStreamCondition(
		d.Get("project_name").(string),
		d.Id(),
		attrs,
	)

	return err
}

func resourceStreamConditionImport(d *schema.ResourceData, m interface{}) ([]*schema.ResourceData, error) {
	client := m.(*lightstep.Client)

	ids := strings.Split(d.Id(), ".")
	if len(ids) != 2 {
		return []*schema.ResourceData{}, fmt.Errorf("Error importing lightstep_condition. Expecting an  ID formed as '<lightstep_project>.<lightstep_condition_ID>'")
	}

	project, id := ids[0], ids[1]
	c, err := client.GetStreamCondition(project, id)
	if err != nil {
		return []*schema.ResourceData{}, err
	}

	// stream ID does not get returned from getCondition
	// need to follow the links in relationships to get stream ID
	stream_id, err := client.GetStreamIDByLink(c.Relationships.Stream.Links.Related)
	if err != nil {
		return []*schema.ResourceData{}, err
	}

	d.SetId(id)
	d.Set("project_name", project)                                 //nolint these are values fetched from the api
	d.Set("condition_name", c.Attributes.Name)                     //nolint so we know that they are valid
	d.Set("expression", c.Attributes.Expression)                   //nolint
	d.Set("evaluation_window_ms", c.Attributes.EvaluationWindowMS) //nolint
	d.Set("stream_id", stream_id)                                  //nolint

	return []*schema.ResourceData{d}, nil
}