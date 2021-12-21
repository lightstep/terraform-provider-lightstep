package lightstep

import (
	"context"
	"fmt"
	"github.com/lightstep/terraform-provider-lightstep/client"
	"net/http"
	"strings"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func resourceStreamCondition() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceStreamConditionCreate,
		ReadContext:   resourceStreamConditionRead,
		DeleteContext: resourceStreamConditionDelete,
		UpdateContext: resourceStreamConditionUpdate,
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
				ForceNew: true,
			},
			"evaluation_window_ms": {
				Type:     schema.TypeInt,
				Required: true,
			},
		},
	}
}

func resourceStreamConditionCreate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	var diags diag.Diagnostics

	c := m.(*client.Client)

	condition, err := c.CreateStreamCondition(
		d.Get("project_name").(string),
		d.Get("condition_name").(string),
		d.Get("expression").(string),
		d.Get("evaluation_window_ms").(int),
		d.Get("stream_id").(string))
	if err != nil {
		return diag.FromErr(fmt.Errorf("Failed to create stream condition: %v", err))
	}

	d.SetId(condition.ID)
	if err := setResourceDataFromStreamCondition(d, condition); err != nil {
		return diag.FromErr(fmt.Errorf("Failed to set stream condition response from API to terraform state: %v", err))
	}

	return diags
}

func resourceStreamConditionRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	var diags diag.Diagnostics

	c := m.(*client.Client)
	condition, err := c.GetStreamCondition(d.Get("project_name").(string), d.Id())
	if err != nil {
		apiErr := err.(client.APIResponseCarrier)
		if apiErr.GetHTTPResponse().StatusCode == http.StatusNotFound {
			d.SetId("")
			return diags
		}
		return diag.FromErr(fmt.Errorf("Failed to get stream condition: %v\n", apiErr))
	}

	if err := setResourceDataFromStreamCondition(d, *condition); err != nil {
		return diag.FromErr(fmt.Errorf("Failed to set stream condition response from API to terraform state: %v", err))
	}

	return diags
}

func resourceStreamConditionDelete(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	var diags diag.Diagnostics

	c := m.(*client.Client)
	if err := c.DeleteStreamCondition(d.Get("project_name").(string), d.Id()); err != nil {
		return diag.FromErr(fmt.Errorf("Failed to delete stream condition: %v", err))
	}

	return diags
}

func resourceStreamConditionUpdate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	var diags diag.Diagnostics

	c := m.(*client.Client)
	attrs := client.StreamConditionAttributes{
		Name:               d.Get("condition_name").(string),
		EvaluationWindowMS: d.Get("evaluation_window_ms").(int),
		Expression:         d.Get("expression").(string),
	}

	condition, err := c.UpdateStreamCondition(d.Get("project_name").(string), d.Id(), attrs)
	if err != nil {
		return diag.FromErr(fmt.Errorf("Failed to update stream condition: %v", err))
	}

	if err := setResourceDataFromStreamCondition(d, *condition); err != nil {
		return diag.FromErr(fmt.Errorf("Failed to set stream condition from API response to terraform state: %v", err))
	}

	return diags
}

func resourceStreamConditionImport(d *schema.ResourceData, m interface{}) ([]*schema.ResourceData, error) {
	c := m.(*client.Client)

	ids := strings.Split(d.Id(), ".")
	if len(ids) != 2 {
		return []*schema.ResourceData{}, fmt.Errorf("Error importing lightstep_condition. Expecting an  ID formed as '<lightstep_project>.<lightstep_condition_ID>'")
	}

	project, id := ids[0], ids[1]
	condition, err := c.GetStreamCondition(project, id)
	if err != nil {
		return []*schema.ResourceData{}, err
	}

	d.SetId(id)
	if err := d.Set("project_name", project); err != nil {
		return []*schema.ResourceData{}, nil
	}

	if err := setResourceDataFromStreamCondition(d, *condition); err != nil {
		return []*schema.ResourceData{d}, err
	}

	return []*schema.ResourceData{d}, nil
}

// update terraform state with stream condition API call response
func setResourceDataFromStreamCondition(d *schema.ResourceData, sc client.StreamCondition) error {
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