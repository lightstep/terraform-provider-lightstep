package main

import (
	"fmt"
	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
	"github.com/lightstep/terraform-provider-lightstep/lightstep"
	"strings"
)

func resourceCondition() *schema.Resource {
	return &schema.Resource{
		Create: resourceConditionCreate,
		Read:   resourceConditionRead,
		Delete: resourceConditionDelete,
		Update: resourceConditionUpdate,
		Importer: &schema.ResourceImporter{
			State: resourceConditionImport,
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

func resourceConditionCreate(d *schema.ResourceData, m interface{}) error {
	client := m.(*lightstep.Client)

	condition, err := client.CreateCondition(
		d.Get("project_name").(string),
		d.Get("condition_name").(string),
		d.Get("expression").(string),
		d.Get("evaluation_window_ms").(int),
		d.Get("stream_id").(string))

	if err != nil {
		return err
	}
	d.SetId(condition.ID)
	return resourceConditionRead(d, m)
}

func resourceConditionRead(d *schema.ResourceData, m interface{}) error {
	client := m.(*lightstep.Client)
	_, err := client.GetCondition(d.Get("project_name").(string), d.Id())
	if err != nil {
		return err
	}
	return nil
}

func resourceConditionDelete(d *schema.ResourceData, m interface{}) error {
	client := m.(*lightstep.Client)
	err := client.DeleteCondition(d.Get("project_name").(string), d.Id())
	return err
}

func resourceConditionUpdate(d *schema.ResourceData, m interface{}) error {
	client := m.(*lightstep.Client)

	attrs := lightstep.ConditionAttributes{
		Name:               d.Get("condition_name").(string),
		EvaluationWindowMS: d.Get("evaluation_window_ms").(int),
		Expression:         d.Get("expression").(string),
	}

	_, err := client.UpdateCondition(
		d.Get("project_name").(string),
		d.Id(),
		attrs,
	)

	return err
}

func resourceConditionImport(d *schema.ResourceData, m interface{}) ([]*schema.ResourceData, error) {
	client := m.(*lightstep.Client)

	ids := strings.Split(d.Id(), ".")
	if len(ids) != 2 {
		return []*schema.ResourceData{}, fmt.Errorf("Error importing lightstep_condition. Expecting an  ID formed as '<lightstep_project>.<lightstep_condition>'")
	}
	project, id := ids[0], ids[1]

	_, err := client.GetCondition(project, id)
	if err != nil {
		return []*schema.ResourceData{}, err
	}

	d.SetId(id)
	d.Set("project", project)

	return []*schema.ResourceData{d}, nil
}
