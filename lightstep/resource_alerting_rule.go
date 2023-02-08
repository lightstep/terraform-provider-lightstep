package lightstep

import (
	"context"
	"fmt"
	"net/http"
	"strings"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/lightstep/terraform-provider-lightstep/client"
)

func resourceAlertingRule() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceAlertingRuleCreate,
		ReadContext:   resourceAlertingRuleRead,
		DeleteContext: resourceAlertingRuleDelete,
		Importer: &schema.ResourceImporter{
			StateContext: resourceAlertingRuleImport,
		},
		Schema: map[string]*schema.Schema{
			"project_name": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true, // API does not provide an Update method
			},
			"condition_id": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true, // API does not provide an Update method
			},
			"destination_id": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true, // API does not provide an Update method
			},
			"update_interval": {
				Type:         schema.TypeString,
				ValidateFunc: validation.StringInSlice(GetValidUpdateInterval(), false),
				Required:     true,
				ForceNew:     true, // API does not provide an Update method
			},
		},
	}
}

func resourceAlertingRuleCreate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	var diags diag.Diagnostics

	client := m.(*client.Client)

	updateIntervalMS := validUpdateInterval[d.Get("update_interval").(string)]

	rule, err := client.CreateAlertingRule(
		ctx,
		d.Get("project_name").(string),
		updateIntervalMS,
		d.Get("destination_id").(string),
		d.Get("condition_id").(string))
	if err != nil {
		return diag.FromErr(fmt.Errorf("failed to create alerting rule: %v", err))
	}

	d.SetId(rule.ID)
	if err := setResourceDataFromAlertingRule(d, rule); err != nil {
		return diag.FromErr(fmt.Errorf("failed to set alerting rule response from API to terraform state: %v", err))
	}

	return diags
}

func resourceAlertingRuleRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	var diags diag.Diagnostics

	c := m.(*client.Client)
	rule, err := c.GetAlertingRule(ctx, d.Get("project_name").(string), d.Id())
	if err != nil {
		apiErr, ok := err.(client.APIResponseCarrier)

		if !ok {
			return diag.FromErr(fmt.Errorf("failed to get alerting rule: %v", err))
		}

		if apiErr.GetStatusCode() == http.StatusNotFound {
			d.SetId("")
			return diags
		}

		return diag.FromErr(fmt.Errorf("failed to get alerting rule: %v", err))
	}

	if err := setResourceDataFromAlertingRule(d, *rule); err != nil {
		return diag.FromErr(fmt.Errorf("failed to set alerting rule response from API to terraform state: %v", err))
	}

	return diags
}

func resourceAlertingRuleDelete(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	var diags diag.Diagnostics

	client := m.(*client.Client)
	if err := client.DeleteAlertingRule(ctx, d.Get("project_name").(string), d.Id()); err != nil {
		return diag.FromErr(fmt.Errorf("failed to delete alerting rule: %v", err))
	}

	return diags
}

func resourceAlertingRuleImport(ctx context.Context, d *schema.ResourceData, m interface{}) ([]*schema.ResourceData, error) {
	client := m.(*client.Client)

	ids := strings.Split(d.Id(), ".")
	if len(ids) != 2 {
		return []*schema.ResourceData{}, fmt.Errorf("error importing lightstep_alerting_rule. Expecting an  ID formed as '<lightstep_project>.<lightstep_lightstep_alerting_rule_ID>'")
	}

	project, id := ids[0], ids[1]
	rule, err := client.GetAlertingRule(ctx, project, id)
	if err != nil {
		return []*schema.ResourceData{}, err
	}

	d.SetId(id)
	if err := d.Set("project_name", project); err != nil {
		return []*schema.ResourceData{}, nil
	}

	if err := setResourceDataFromAlertingRule(d, *rule); err != nil {
		return []*schema.ResourceData{d}, err
	}

	return []*schema.ResourceData{d}, nil
}

// update terraform state with alerting rule API call response
func setResourceDataFromAlertingRule(d *schema.ResourceData, rule client.StreamAlertingRuleResponse) error {
	if err := d.Set("condition_id", rule.Relationships.Condition.Data.ID); err != nil {
		return err
	}

	if err := d.Set("destination_id", rule.Relationships.Destination.Data.ID); err != nil {
		return err
	}

	if err := d.Set("update_interval", GetUpdateIntervalValue(rule.Attributes.UpdateInterval)); err != nil {
		return err
	}

	return nil
}
