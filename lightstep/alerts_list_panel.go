package lightstep

import (
	"fmt"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"

	"github.com/lightstep/terraform-provider-lightstep/client"
)

const (
	AlertsListPanel     = "alerts_list_panel"
	AlertsListType      = "alerts_list"
	AlertsListPanelDesc = "A dashboard panel to view a list of your alerts and their status"
)

func getAlertsListPanelSchema() *schema.Schema {
	elements := mergeSchemas(
		getPositionSchema(),
		map[string]*schema.Schema{
			"name": {
				Type:     schema.TypeString,
				Optional: true,
				Default:  "Alerts List Panel",
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
								"status",
								"name",
								"labels",
								"snooze",
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
					},
				},
			},
			"filter_by": {
				Type:        schema.TypeSet,
				Optional:    true,
				Description: "a list of predicates that are implicitly ANDed together to filter alerts",
				MaxItems:    1,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"predicate": {
							Type:        schema.TypeSet,
							Optional:    true,
							Description: "a single predicate",
							Elem: &schema.Resource{
								Schema: getPredicateSchemaMap(),
							},
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
		Description: AlertsListPanelDesc,
		Elem:        &schema.Resource{Schema: elements},
	}
}

func getPredicateSchemaMap() map[string]*schema.Schema {
	return map[string]*schema.Schema{
		"operator": {
			Type:     schema.TypeString,
			Optional: true,
			ValidateFunc: validation.StringInSlice([]string{
				"eq",
				"neq",
			}, false),
		},
		"label": {
			Type:        schema.TypeSet,
			Optional:    true,
			Description: "Labels can be key/value pairs or standalone values.",
			Elem: &schema.Resource{
				Schema: getLabelsSchemaMap(),
			},
		},
	}
}

func getLabelsSchemaMap() map[string]*schema.Schema {
	return map[string]*schema.Schema{
		"key": {
			Type:     schema.TypeString,
			Optional: true,
		},
		"value": {
			Type:     schema.TypeString,
			Required: true,
		},
	}
}

func convertAlertsListFromResourceToApiRequest(alertsListPanelsIn interface{}) ([]client.Panel, error) {
	in := alertsListPanelsIn.(*schema.Set).List()
	var alertsListPanels []client.Panel

	for _, s := range in {
		alertsListPanel, ok := s.(map[string]interface{})
		if !ok {
			return nil, fmt.Errorf("bad format, %v", s)
		}
		p := client.Panel{
			ID:       alertsListPanel["id"].(string),
			Title:    alertsListPanel["name"].(string),
			Type:     AlertsListType,
			Position: buildPosition(alertsListPanel),
		}

		p.Body = map[string]interface{}{}
		// N.B. panel_options are optional, so we don't return an error if not found
		if opts, ok := alertsListPanel["panel_options"].(*schema.Set); ok {
			list := opts.List()
			count := len(list)
			if count > 1 {
				return nil, fmt.Errorf("panel_options must be defined only once")
			} else if count == 1 {
				resourceDisplayOptions, ok := list[0].(map[string]interface{})
				if !ok {
					return nil, fmt.Errorf("unexpected format for panel_options")
				}
				for k, v := range resourceDisplayOptions {
					// don't want to send 'sort_by: ""' to the API etc.
					if v == "" {
						delete(resourceDisplayOptions, k)
					}
				}
				if len(resourceDisplayOptions) > 0 {
					p.Body["display_options"] = resourceDisplayOptions
				}
			}
		}
		if opts, ok := alertsListPanel["filter_by"].(*schema.Set); ok {
			list := opts.List()
			count := len(list)
			if count > 1 {
				return nil, fmt.Errorf("filter_by must be defined only once")
			} else if count == 1 {
				filterByResource, ok := list[0].(map[string]interface{})
				filterBy := make(map[string]interface{})
				if !ok {
					return nil, fmt.Errorf("unexpected format for filter_by")
				}
				if maybePredicateResources, ok := filterByResource["predicate"]; ok {
					predicateResources, ok := maybePredicateResources.(*schema.Set)
					if !ok {
						return nil, fmt.Errorf("unexpected format for predicate")
					}
					var predicates []map[string]interface{}
					for _, p := range predicateResources.List() {
						predicateResource, ok := p.(map[string]interface{})
						if !ok {
							return nil, fmt.Errorf("unexpected format for predicate")
						}
						predicate := make(map[string]interface{})
						predicate["operator"] = predicateResource["operator"]
						maybeLabelResources := predicateResource["label"]
						if labelResources, ok := maybeLabelResources.(*schema.Set); ok {
							var labels []map[string]interface{}
							for _, l := range labelResources.List() {
								label, ok := l.(map[string]interface{})
								if !ok {
									return nil, fmt.Errorf("unexpected format for label")
								}
								labels = append(labels, label)
							}
							predicate["labels"] = labels
						}
						predicates = append(predicates, predicate)
					}
					filterBy["predicates"] = predicates
				}

				p.Body["filter_by"] = filterBy
			}
		}
		alertsListPanels = append(alertsListPanels, p)
	}

	return alertsListPanels, nil
}

func convertAlertsListFromApiRequestToResource(apiPanels []client.Panel) []interface{} {
	var alertsListPanelResources []interface{}
	for _, p := range apiPanels {
		if p.Type == AlertsListType {
			resource := map[string]interface{}{}
			// Alias for what we refer to as title elsewhere
			resource["name"] = p.Title
			resource["x_pos"] = p.Position.XPos
			resource["y_pos"] = p.Position.YPos
			resource["width"] = p.Position.Width
			resource["height"] = p.Position.Height
			resource["id"] = p.ID
			// N.B. the panel body might be nil. panel_options are optional for the alerts list panel.
			if p.Body != nil {
				if maybeDisplayOptions, ok := p.Body["display_options"]; ok {
					displayOptions, ok := maybeDisplayOptions.(map[string]interface{})
					if ok {
						resource["panel_options"] = convertNestedMapToSchemaSet(displayOptions)
					}
				}
				if maybeFilterBy, ok := p.Body["filter_by"]; ok {
					filterBy, ok := maybeFilterBy.(map[string]interface{})
					if ok {
						filterByResource := make(map[string]interface{})
						if maybePredicates, ok := filterBy["predicates"]; ok {
							predicates := maybePredicates.([]interface{})
							var predicateResources []interface{}
							for _, p := range predicates {
								predicate, ok := p.(map[string]interface{})
								if ok {
									predicateResource := make(map[string]interface{})
									if maybeOperator, ok := predicate["operator"]; ok {
										operator := maybeOperator.(string)
										predicateResource["operator"] = operator
									}
									if maybeLabels, ok := predicate["labels"]; ok {
										labels := maybeLabels.([]interface{})
										predicateResource["label"] = schema.NewSet(
											schema.HashResource(&schema.Resource{
												Schema: getLabelsSchemaMap(),
											}),
											labels)
									}
									predicateResources = append(predicateResources, predicateResource)
									filterByResource["predicate"] = schema.NewSet(
										schema.HashResource(&schema.Resource{
											Schema: getPredicateSchemaMap(),
										}),
										predicateResources)
								}
							}
						}
						resource["filter_by"] = convertNestedMapToSchemaSet(filterByResource)
					}
				}
			}
			alertsListPanelResources = append(alertsListPanelResources, resource)
		}
	}
	return alertsListPanelResources
}
