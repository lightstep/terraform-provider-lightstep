package main

import (
	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
	"github.com/lightstep/terraform-provider-lightstep/lightstep"
)

func resourceCondition() *schema.Resource {
	return &schema.Resource{
		Create: resourceConditionCreate,
		Read:   resourceConditionRead,
		Delete: resourceConditionDelete,
		Update: resourceConditionUpdate,

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
	err := client.DeleteCondition(d.Get("project_name").(string), d.Get("condition_id").(string))
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
