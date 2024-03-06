package lightstep

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"strings"

	"github.com/lightstep/terraform-provider-lightstep/client"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func resourceEventQuery() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceEventQueryCreate,
		ReadContext:   resourceEventQueryRead,
		UpdateContext: resourceEventQueryUpdate,
		DeleteContext: resourceEventQueryDelete,
		Importer: &schema.ResourceImporter{
			StateContext: resourceEventQueryImport,
		},
		Schema: map[string]*schema.Schema{
			"project_name": {
				Type:        schema.TypeString,
				Required:    true,
				ForceNew:    true,
				Description: "Lightstep project name",
			},
			"name": {
				Type:     schema.TypeString,
				Required: true,
			},
			"source": {
				Type:     schema.TypeString,
				Required: true,
			},
			"query_string": {
				Type:     schema.TypeString,
				Required: true,
			},
			"type": {
				Type:     schema.TypeString,
				Required: true,
			},
			"description": {
				Type:     schema.TypeString,
				Optional: true,
			},
			"tooltip_fields": {
				Type:     schema.TypeList,
				Optional: true,
				Elem: &schema.Schema{
					Type: schema.TypeString,
				},
			},
		},
	}
}

func resourceEventQueryRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	c := m.(*client.Client)

	eq, err := c.GetEventQuery(ctx, d.Get("project_name").(string), d.Id())
	if err != nil {
		apiErr, ok := err.(client.APIResponseCarrier)
		if !ok {
			return diag.FromErr(fmt.Errorf("failed to get event query: %v", err))
		}
		if apiErr.GetStatusCode() == http.StatusNotFound {
			d.SetId("")
			return diags
		}
		return diag.FromErr(fmt.Errorf("failed to get event query: %v", apiErr))
	}

	if err := d.Set("name", eq.Name); err != nil {
		return diag.FromErr(fmt.Errorf("unable to set query name: %v", err))
	}
	if err := d.Set("type", eq.Type); err != nil {
		return diag.FromErr(fmt.Errorf("unable to set query type: %v", err))
	}
	if err := d.Set("query_string", eq.QueryString); err != nil {
		return diag.FromErr(fmt.Errorf("unable to set query string: %v", err))
	}
	if err := d.Set("source", eq.Source); err != nil {
		return diag.FromErr(fmt.Errorf("unable to set query string: %v", err))
	}
	if err := d.Set("description", eq.Description); err != nil {
		return diag.FromErr(fmt.Errorf("unable to set description: %v", err))
	}
	if err := d.Set("tooltip_fields", eq.TooltipFields); err != nil {
		return diag.FromErr(fmt.Errorf("unable to set tooltip fields: %v", err))
	}

	return diags
}

func resourceDataToStringSlice(resourceData *schema.ResourceData, fieldName string) []string {
	resource := resourceData.Get(fieldName)

	asInterfaceSlice := resource.([]interface{})
	stringSlice := make([]string, len(asInterfaceSlice))
	for i, element := range asInterfaceSlice {
		stringSlice[i] = element.(string)
	}
	return stringSlice
}

func resourceEventQueryCreate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	c := m.(*client.Client)

	attrs := client.EventQueryAttributes{
		Type:          d.Get("type").(string),
		Name:          d.Get("name").(string),
		Source:        d.Get("source").(string),
		QueryString:   d.Get("query_string").(string),
		Description:   d.Get("description").(string),
		TooltipFields: resourceDataToStringSlice(d, "tooltip_fields"),
	}
	eq, err := c.CreateEventQuery(ctx, d.Get("project_name").(string), attrs)
	if err != nil {
		return diag.FromErr(fmt.Errorf("failed to create event query %v: %v", attrs.Name, err))
	}

	d.SetId(eq.ID)
	return resourceEventQueryRead(ctx, d, m)
}

func resourceEventQueryImport(ctx context.Context, d *schema.ResourceData, m interface{}) ([]*schema.ResourceData, error) {
	c := m.(*client.Client)

	ids := strings.Split(d.Id(), ".")
	if len(ids) != 2 {
		return []*schema.ResourceData{}, errors.New("error importing event query. Expecting an  ID formed as '<lightstep_project>.<event_query_ID>'")
	}

	project, id := ids[0], ids[1]
	eq, err := c.GetEventQuery(ctx, project, id)
	if err != nil {
		return []*schema.ResourceData{}, fmt.Errorf("failed to get event query: %v", err)
	}

	d.SetId(eq.ID)
	if err := d.Set("project_name", project); err != nil {
		return []*schema.ResourceData{}, fmt.Errorf("unable to set project_name resource field: %v", err)
	}
	if err := d.Set("name", eq.Name); err != nil {
		return nil, fmt.Errorf("unable to set query name: %v", err)
	}
	if err := d.Set("type", eq.Type); err != nil {
		return nil, fmt.Errorf("unable to set query type: %v", err)
	}
	if err := d.Set("query_string", eq.QueryString); err != nil {
		return nil, fmt.Errorf("unable to set query string: %v", err)
	}
	if err := d.Set("source", eq.Source); err != nil {
		return nil, fmt.Errorf("unable to set query string: %v", err)
	}
	if err := d.Set("description", eq.Description); err != nil {
		return nil, fmt.Errorf("unable to set description: %v", err)
	}
	if err := d.Set("tooltip_fields", eq.TooltipFields); err != nil {
		return nil, fmt.Errorf("unable to set tooltip fields: %v", err)
	}
	return []*schema.ResourceData{d}, nil
}

func resourceEventQueryUpdate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	c := m.(*client.Client)

	attrs := client.EventQueryAttributes{
		Type:          d.Get("type").(string),
		Name:          d.Get("name").(string),
		Source:        d.Get("source").(string),
		QueryString:   d.Get("query_string").(string),
		Description:   d.Get("description").(string),
		TooltipFields: resourceDataToStringSlice(d, "tooltip_fields"),
	}
	eq, err := c.UpdateEventQuery(ctx, d.Get("project_name").(string), d.Id(), attrs)
	if err != nil {
		return diag.FromErr(fmt.Errorf("failed to create event query %v: %v", attrs.Name, err))
	}

	d.SetId(eq.ID)
	return resourceEventQueryRead(ctx, d, m)
}

func resourceEventQueryDelete(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	var diags diag.Diagnostics

	client := m.(*client.Client)
	if err := client.DeleteEventQuery(ctx, d.Get("project_name").(string), d.Id()); err != nil {
		return diag.FromErr(fmt.Errorf("failed to delete event query: %v", err))
	}

	return diags
}
