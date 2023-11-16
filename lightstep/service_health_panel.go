package lightstep

import (
	"fmt"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"

	"github.com/lightstep/terraform-provider-lightstep/client"
)

const (
	ServiceHealthPanel     = "service_health_panel"
	ServiceHealthType      = "service_health"
	ServiceHealthPanelDesc = "A dashboard panel to view the health of your services"
)

func getServiceHealthPanelSchema() *schema.Schema {
	elements := mergeSchemas(
		getPositionSchema(),
		map[string]*schema.Schema{
			"name": {
				Type:     schema.TypeString,
				Optional: true,
				Default:  "Service Health Panel",
			},
			"id": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"panel_options": {
				Type:        schema.TypeSet,
				Optional:    true,
				Description: "custom options for the service health panel",
				MaxItems:    1,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"sort_by": {
							Type:     schema.TypeString,
							Optional: true,
							ValidateFunc: validation.StringInSlice([]string{
								"service",
								"latency",
								"error",
								"rate",
							}, false),
						},
						"sort_direction": {
							Type:     schema.TypeString,
							Optional: true,
							ValidateFunc: validation.StringInSlice([]string{
								"asc",
								"desc",
								"error",
								"rate",
							}, false),
						},
						"percentile": {
							Type:     schema.TypeString,
							Optional: true,
							ValidateFunc: validation.StringInSlice([]string{
								"p50",
								"p90",
								"p95",
								"p99",
							}, false),
						},
						"change_since": {
							Type:     schema.TypeString,
							Optional: true,
							ValidateFunc: validation.StringInSlice([]string{
								"1h",
								"1d",
								"3d",
								"p99",
							}, false),
						},
					},
				},
			},
		},
	)

	return &schema.Schema{
		Type:        schema.TypeSet,
		Optional:    true,
		Computed:    true, // the panels can be mutated individually; panel mutations should not trigger group updates
		Description: ServiceHealthPanelDesc,
		Elem:        &schema.Resource{Schema: elements},
	}
}

func convertServiceHealthFromResourceToApiRequest(serviceHealthPanelsIn interface{}) ([]client.Panel, error) {
	in := serviceHealthPanelsIn.(*schema.Set).List()
	var serviceHealthPanels []client.Panel

	for _, s := range in {
		serviceHealthPanel, ok := s.(map[string]interface{})
		if !ok {
			return nil, fmt.Errorf("bad format, %v", s)
		}
		p := client.Panel{
			ID:       serviceHealthPanel["id"].(string),
			Title:    serviceHealthPanel["name"].(string),
			Type:     ServiceHealthType,
			Position: buildPosition(serviceHealthPanel),
		}

		p.Body = map[string]interface{}{}
		//N.B. panel_options are optional, so we don't return an error if not found
		if opts, ok := serviceHealthPanel["panel_options"].(*schema.Set); ok {
			list := opts.List()
			count := len(list)
			if count > 1 {
				return nil, fmt.Errorf("panel_options must be defined only once")
			} else if count == 1 {
				m, ok := list[0].(map[string]interface{})
				if !ok {
					return nil, fmt.Errorf("unexpected format for panel_options")
				}
				// The API treats panel_options as an opaque blob, so we can pass what we have along directly
				p.Body["display_options"] = m
			}
		}

		serviceHealthPanels = append(serviceHealthPanels, p)
	}

	return serviceHealthPanels, nil
}

func convertServiceHealthfromApiRequestToResource(apiPanels []client.Panel) []interface{} {
	var serviceHealthPanelResources []interface{}
	for _, p := range apiPanels {
		if p.Type == ServiceHealthType {
			resource := map[string]interface{}{}
			// Alias for what we refer to as title elsewhere
			resource["name"] = p.Title
			resource["x_pos"] = p.Position.XPos
			resource["y_pos"] = p.Position.YPos
			resource["width"] = p.Position.Width
			resource["height"] = p.Position.Height
			resource["id"] = p.ID
			// N.B. the panel body might be nil. panel_options are optional for the service health panel.
			if p.Body != nil {
				if maybeDisplayOptions, ok := p.Body["display_options"]; ok {
					displayOptions, ok := maybeDisplayOptions.(map[string]interface{})
					if ok {
						resource["panel_options"] = convertNestedMapToSchemaSet(displayOptions)
					}
				}
			}
			serviceHealthPanelResources = append(serviceHealthPanelResources, resource)
		}
	}
	return serviceHealthPanelResources
}
