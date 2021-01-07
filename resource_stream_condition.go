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
		d.Get("stream_id").(string),
	)
	if err != nil {
		return err
	}

	d.SetId(condition.ID)
	return nil
}

func resourceStreamConditionRead(d *schema.ResourceData, m interface{}) error {
	client := m.(*lightstep.Client)
	projectName := d.Get("project_name").(string)
	streamCond, err := client.GetStreamCondition(projectName, d.Id())
	if err != nil {
		return err
	}

	if err := readResourceDataFromStreamCondition(d, streamCond); err != nil {
		return err
	}

	return nil
}

func readResourceDataFromStreamCondition(d *schema.ResourceData, sc lightstep.StreamCondition) error {
	if err := d.Set("condition_name", sc.Attributes.Name); err != nil {
		return err
	}

	if err := d.Set("expression", sc.Attributes.Expression); err != nil {
		return err
	}

	if err := d.Set("evaluation_window_ms", sc.Attributes.EvaluationWindowMS); err != nil {
		return err
	}

	rel := strings.Split(sc.Relationships.Stream.Links.Related, "/")
	streamID := rel[len(rel)-1]

	if err := d.Set("stream_id", streamID); err != nil {
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
	streamID, err := client.GetStreamIDByLink(c.Relationships.Stream.Links.Related)
	if err != nil {
		return []*schema.ResourceData{}, err
	}

	d.SetId(id)

	if err := d.Set("project_name", project); err != nil {
		return []*schema.ResourceData{}, nil
	}

	if err := d.Set("condition_name", c.Attributes.Name); err != nil {
		return []*schema.ResourceData{}, nil
	}

	if err := d.Set("expression", c.Attributes.Expression); err != nil {
		return []*schema.ResourceData{}, nil
	}

	if err := d.Set("evaluation_window_ms", c.Attributes.EvaluationWindowMS); err != nil {
		return []*schema.ResourceData{}, nil
	}

	if err := d.Set("stream_id", streamID); err != nil {
		return []*schema.ResourceData{}, nil
	}

	return []*schema.ResourceData{d}, nil
}
