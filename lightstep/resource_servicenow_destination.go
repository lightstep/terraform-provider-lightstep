package lightstep

import (
	"context"
	"fmt"
	"strings"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"

	"github.com/lightstep/terraform-provider-lightstep/client"
)

func resourceServiceNowDestination() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceServiceNowDestinationCreate,
		ReadContext:   resourceDestinationRead,
		DeleteContext: resourceDestinationDelete,
		Importer: &schema.ResourceImporter{
			StateContext: resourceServiceNowDestinationImport,
		},
		Schema: map[string]*schema.Schema{
			"project_name": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"destination_name": {
				Type:        schema.TypeString,
				Required:    true,
				ForceNew:    true,
				Description: "Name of the ServiceNow destination",
			},
			"url": {
				Type:         schema.TypeString,
				Required:     true,
				ForceNew:     true,
				Description:  "ServiceNow instance URL",
				ValidateFunc: validation.IsURLWithScheme([]string{"https"}),
			},
			"auth": {
				Type:        schema.TypeList,
				MinItems:    1,
				MaxItems:    1,
				Required:    true,
				ForceNew:    true,
				Description: "Basic auth used to authenticate with the ServiceNow instance",
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"username": {
							Type:     schema.TypeString,
							Required: true,
							ForceNew: true,
						},
						"password": {
							Type:      schema.TypeString,
							Sensitive: true,
							Required:  true,
							ForceNew:  true,
						},
					},
				},
			},
		},
	}
}

func resourceServiceNowDestinationCreate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	c := m.(*client.Client)
	attrs := client.ServiceNowAttributes{
		Name:            d.Get("destination_name").(string),
		DestinationType: "servicenow",
		URL:             d.Get("url").(string),
	}
	auth := d.Get("auth").([]interface{})[0].(map[string]interface{})
	attrs.Auth = client.Auth{
		Username: auth["username"].(string),
		Password: auth["password"].(string),
	}
	dest := client.Destination{
		Type:       "destination",
		Attributes: attrs,
	}

	destination, err := c.CreateDestination(ctx, d.Get("project_name").(string), dest)
	if err != nil {
		return diag.FromErr(fmt.Errorf("failed to create ServiceNow destination %v: %v", attrs.Name, err))
	}

	d.SetId(destination.ID)
	return resourceDestinationRead(ctx, d, m)
}

func resourceServiceNowDestinationImport(ctx context.Context, d *schema.ResourceData, m interface{}) ([]*schema.ResourceData, error) {
	c := m.(*client.Client)

	ids := strings.Split(d.Id(), ".")
	if len(ids) != 2 {
		return []*schema.ResourceData{}, fmt.Errorf("error importing lightstep_servicenow_destination. Expecting an  ID formed as '<lightstep_project>.<lightstep_destination_ID>'")
	}

	project, id := ids[0], ids[1]
	dest, err := c.GetDestination(ctx, project, id)
	if err != nil {
		return []*schema.ResourceData{}, fmt.Errorf("failed to get ServiceNow destination: %v", err)
	}

	d.SetId(dest.ID)
	if err := d.Set("project_name", project); err != nil {
		return []*schema.ResourceData{}, fmt.Errorf("unable to set project_name resource field: %v", err)
	}

	attributes := dest.Attributes.(map[string]interface{})
	if err := d.Set("destination_name", attributes["name"]); err != nil {
		return []*schema.ResourceData{}, fmt.Errorf("unable to set destination_name resource field: %v", err)
	}

	if err := d.Set("url", attributes["url"]); err != nil {
		return []*schema.ResourceData{}, fmt.Errorf("unable to set url resource field: %v", err)
	}

	if err := d.Set("auth", []interface{}{attributes["auth"]}); err != nil {
		return []*schema.ResourceData{}, fmt.Errorf("unable to set auth resource field: %v", err)
	}

	return []*schema.ResourceData{d}, nil
}
