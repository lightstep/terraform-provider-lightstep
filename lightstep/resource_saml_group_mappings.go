package lightstep

import (
	"context"
	"errors"
	"fmt"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"regexp"

	"github.com/lightstep/terraform-provider-lightstep/client"
)

func resourceSAMLGroupMappings() *schema.Resource {
	return &schema.Resource{
		Description:   `Provides a Lightstep SAML Group Mapping to automatically update user's roles based on their SAML attributes. For conceptual information about managing SAML group mappings, visit [Lightstep's documentation](https://docs.lightstep.com/docs/map-saml-attributes).`,
		CreateContext: resourceSAMLGroupMappingsCreateOrUpdate,
		ReadContext:   resourceSAMLGroupMappingsRead,
		UpdateContext: resourceSAMLGroupMappingsCreateOrUpdate,
		DeleteContext: resourceSAMLGroupMappingsDelete,
		Schema: map[string]*schema.Schema{
			"mapping": {
				Type:        schema.TypeSet,
				Required:    true,
				Description: "List of SAML Group Mappings.",
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"match": {
							Type:        schema.TypeList,
							Required:    true,
							MaxItems:    1,
							Description: "Key/Value attribute pair to match against the user's SAML attributes.",
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"attribute_key": {
										Type:        schema.TypeString,
										Required:    true,
										Description: "Attribute Key to match against the user's SAML attributes.",
									},
									"attribute_value": {
										Type:        schema.TypeString,
										Required:    true,
										Description: "Attribute Value to match against the user's SAML attributes",
									},
								},
							},
						},
						"roles": {
							Type:        schema.TypeList,
							Required:    true,
							MaxItems:    1,
							Description: "Roles to assign to the user if the match is successful. ",
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"organization_role": {
										Type:        schema.TypeString,
										Required:    true,
										Description: "Organization Role. Only 'Organization Editor', 'Organization Viewer' and 'Organization Restricted Member'  are supported.",
										ValidateFunc: validation.StringInSlice([]string{
											"Organization Restricted Member",
											"Organization Editor",
											"Organization Viewer",
										}, false),
									},
									"project_roles": {
										Type:        schema.TypeMap,
										Optional:    true,
										Description: "Map of Project Name to Project Role. Only 'Project Editor' and 'Project Viewer' are supported.",
										Elem: &schema.Schema{
											Type: schema.TypeString,
										},
										ValidateDiagFunc: validation.MapValueMatch(
											regexp.MustCompile("Project Editor|Project Viewer"),
											"Project roles must be either 'Project Editor' or 'Project Viewer'"),
									},
								},
							},
						},
					},
				},
			},
		},
	}
}

func resourceSAMLGroupMappingsRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	c := m.(*client.Client)

	samlGroupMappings, err := c.ListSAMLGroupMappings(ctx)
	if err != nil {
		return handleAPIError(err, d, "read SAML group mappings")
	}

	err = setSAMLGroupMappingsResource(d, samlGroupMappings)
	return diag.FromErr(err)
}

func resourceSAMLGroupMappingsCreateOrUpdate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	c := m.(*client.Client)

	// Read the proposed plan.
	mappings, err := getSAMLGroupMappingsResource(d)
	if err != nil {
		return diag.FromErr(err)
	}

	type attribute struct {
		key   string
		value string
	}
	unique := make(map[attribute]bool)
	for _, mapping := range mappings.Mappings {
		attr := attribute{
			key:   mapping.SAMLAttributeKey,
			value: mapping.SAMLAttributeValue,
		}

		if unique[attr] {
			return diag.FromErr(errors.New(fmt.Sprintf("duplicate SAML attribute key/value pair: %s:%s", attr.key, attr.value)))
		}

		unique[attr] = true
	}

	// Update SAML Group Mappings.
	err = c.UpdateSAMLGroupMappings(ctx, mappings)
	if err != nil {
		return handleAPIError(err, d, "create/update SAML group mappings")
	}

	//Update
	err = setSAMLGroupMappingsResource(d, mappings)
	if err != nil {
		return diag.FromErr(err)
	}

	// set a static id.
	d.SetId("saml_group_mappings")

	// Update the state by reading from the API.
	return resourceSAMLGroupMappingsRead(ctx, d, m)
}

func resourceSAMLGroupMappingsDelete(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	c := m.(*client.Client)

	// Update SAML Group Mappings with no mappings.
	err := c.UpdateSAMLGroupMappings(ctx, client.SAMLGroupMappings{})
	if err != nil {
		return handleAPIError(err, d, "delete SAML group mappings")
	}

	// mark this resource as deleted.
	d.SetId("")

	return nil
}

func getSAMLGroupMappingsResource(d *schema.ResourceData) (client.SAMLGroupMappings, error) {
	var mappings client.SAMLGroupMappings

	rawMappings, ok := d.GetOk("mapping")
	if !ok {
		return mappings, nil
	}

	for _, rawMapping := range rawMappings.(*schema.Set).List() {
		var mapping client.SAMLGroupMapping

		rawMappingMap := rawMapping.(map[string]any)

		rawMatch, ok := rawMappingMap["match"]
		if !ok {
			return mappings, errors.New("missing required field 'match'")
		}
		rawMatchSlice := rawMatch.([]any)
		if len(rawMatchSlice) == 0 {
			continue
		}
		matchMap := rawMatchSlice[0].(map[string]any)

		mapping.SAMLAttributeKey = matchMap["attribute_key"].(string)
		mapping.SAMLAttributeValue = matchMap["attribute_value"].(string)

		rawRoles, ok := rawMappingMap["roles"]
		if !ok {
			return mappings, errors.New("missing required field 'roles'")
		}
		rolesMapSlice := rawRoles.([]any)
		if len(rolesMapSlice) == 0 {
			continue
		}
		rolesMap := rolesMapSlice[0].(map[string]any)

		mapping.OrganizationRole = rolesMap["organization_role"].(string)

		mapping.ProjectRoles = make(map[string]string)
		rawProjectRoles := rolesMap["project_roles"].(map[string]any)
		for project, rawRole := range rawProjectRoles {
			mapping.ProjectRoles[project] = rawRole.(string)
		}

		mappings.Mappings = append(mappings.Mappings, mapping)
	}

	return mappings, nil
}

func setSAMLGroupMappingsResource(d *schema.ResourceData, mappings client.SAMLGroupMappings) error {
	var rawMappings []any

	for _, mapping := range mappings.Mappings {
		projectRoles := make(map[string]any)
		for project, role := range mapping.ProjectRoles {
			projectRoles[project] = role
		}

		rawMappings = append(rawMappings, map[string]any{
			"match": []any{
				map[string]any{
					"attribute_key":   mapping.SAMLAttributeKey,
					"attribute_value": mapping.SAMLAttributeValue,
				},
			},
			"roles": []any{
				map[string]any{
					"organization_role": mapping.OrganizationRole,
					"project_roles":     projectRoles,
				},
			},
		})
	}

	return d.Set("mapping", rawMappings)
}
