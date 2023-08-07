package lightstep

import (
	"context"
	"fmt"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/lightstep/terraform-provider-lightstep/client"
	"strings"
)

func resourceInferredServiceRule() *schema.Resource {
	return &schema.Resource{
		Description: "[This resource is under development and is not generally available yet.] " +
			"Provides a Lightstep Inferred Service Rule that can detect and identify inferred services.",
		Schema:        getResourceInferredServiceRuleSchema(),
		CreateContext: resourceInferredServiceRuleCreate,
		ReadContext:   resourceInferredServiceRuleRead,
		UpdateContext: resourceInferredServiceRuleUpdate,
		DeleteContext: resourceInferredServiceRuleDelete,
		Importer:      getResourceInferredServiceRuleImporter(),
	}
}

func getResourceInferredServiceRuleSchema() map[string]*schema.Schema {
	return map[string]*schema.Schema{
		"project_name": {
			Description:  "The name of the project to which the inferred service rule will apply",
			Type:         schema.TypeString,
			Required:     true,
			ValidateFunc: validation.StringIsNotEmpty,
			ForceNew:     true,
		},
		"name": {
			Description:  "The name of the inferred service rule, which is included in the name of each inferred service created from this rule",
			Type:         schema.TypeString,
			Required:     true,
			ValidateFunc: validation.StringLenBetween(0, 190),
		},
		"description": {
			Description:  "An optional description to describe the rule and what services it should infer",
			Type:         schema.TypeString,
			Optional:     true,
			ValidateFunc: validation.StringLenBetween(0, 255),
		},
		"attribute_filters": {
			Description: "Attribute filters that are checked against a leaf span's attributes to indicate the presence of the inferred service",
			Type:        schema.TypeSet,
			Required:    true,
			Elem: &schema.Resource{
				Schema: getResourceInferredServiceRuleAttributeFilterSchema(),
			},
			MinItems: 1,
		},
		"group_by_keys": {
			Description: "A list of attribute keys whose values will be included in the inferred service name",
			Type:        schema.TypeList,
			Optional:    true,
			Elem: &schema.Schema{
				Type: schema.TypeString,
			},
		},
	}
}

func getResourceInferredServiceRuleAttributeFilterSchema() map[string]*schema.Schema {
	return map[string]*schema.Schema{
		"key": {
			Description:  "Key of a span attribute",
			Type:         schema.TypeString,
			Required:     true,
			ValidateFunc: validation.StringIsNotEmpty,
		},
		"values": {
			Description: "Values for the attribute",
			Type:        schema.TypeSet,
			Required:    true,
			Elem: &schema.Schema{
				Type: schema.TypeString,
			},
			MinItems: 1,
		},
	}
}

func resourceInferredServiceRuleCreate(
	ctx context.Context,
	resourceData *schema.ResourceData,
	m interface{},
) diag.Diagnostics {
	apiClient := m.(*client.Client)
	requestAttributes, err := getInferredServiceRuleAttributesFromResource(resourceData)
	if err != nil {
		return diag.FromErr(fmt.Errorf("failed to get inferred service rule request attributes: %v", err))
	}

	created, err := apiClient.CreateInferredServiceRule(ctx, getProjectNameFromResource(resourceData), requestAttributes)
	if err != nil {
		return diag.FromErr(fmt.Errorf("failed to create inferred service rule: %v", err))
	}

	resourceData.SetId(created.ID)
	return resourceInferredServiceRuleRead(ctx, resourceData, m)
}

// resourceInferredServiceRuleRead reads an inferred service rule from the resource data.
//
// When called in as part of Read or Delete, it will read data from the terraform state.
// When called in as part of Create or Update, it will read data from the terraform plan.
func resourceInferredServiceRuleRead(
	ctx context.Context,
	resourceData *schema.ResourceData,
	m interface{},
) diag.Diagnostics {
	var diagnostics diag.Diagnostics

	apiClient := m.(*client.Client)

	projectName := getProjectNameFromResource(resourceData)
	inferredServiceRuleResponse, err := apiClient.GetInferredServiceRule(ctx, projectName, resourceData.Id())
	if err != nil {
		return handleAPIError(err, resourceData, "get inferred service rule")
	}

	err = setResourceDataFromInferredServiceRule(
		projectName,
		&inferredServiceRuleResponse.Attributes,
		resourceData,
	)
	if err != nil {
		diag.Errorf("failed to read inferred service rule due to %v", err)
	}

	return diagnostics
}

func resourceInferredServiceRuleUpdate(
	ctx context.Context,
	resourceData *schema.ResourceData,
	m interface{},
) diag.Diagnostics {
	apiClient := m.(*client.Client)
	requestAttributes, err := getInferredServiceRuleAttributesFromResource(resourceData)
	if err != nil {
		return diag.FromErr(fmt.Errorf("failed to get inferred service rule attributes from resource : %v", err))
	}

	_, err = apiClient.UpdateInferredServiceRule(ctx, getProjectNameFromResource(resourceData), resourceData.Id(), requestAttributes)
	if err != nil {
		return diag.Errorf("failed to update inferred service rule due to error: %v", err)
	}

	return resourceInferredServiceRuleRead(ctx, resourceData, m)
}

func resourceInferredServiceRuleDelete(
	ctx context.Context,
	resourceData *schema.ResourceData,
	m interface{},
) diag.Diagnostics {
	var diagnostics diag.Diagnostics

	apiClient := m.(*client.Client)
	if err := apiClient.DeleteInferredServiceRule(ctx, getProjectNameFromResource(resourceData), resourceData.Id()); err != nil {
		return diag.FromErr(fmt.Errorf("failed to delete inferred service rule: %v", err))
	}

	// resourceData.SetId("") is automatically called assuming delete returns no errors, but is called here to be explicit.
	resourceData.SetId("")
	return diagnostics
}

func getResourceInferredServiceRuleImporter() *schema.ResourceImporter {
	return &schema.ResourceImporter{StateContext: resourceInferredServiceRuleImport}
}

func resourceInferredServiceRuleImport(
	ctx context.Context,
	resourceData *schema.ResourceData,
	m interface{},
) ([]*schema.ResourceData, error) {
	apiClient := m.(*client.Client)

	ids := strings.Split(resourceData.Id(), ".")
	if len(ids) != 2 {
		resourceName := "lightstep_inferred_service_rule"
		return []*schema.ResourceData{}, fmt.Errorf("error importing %v. Expecting a Terraform ID of the form '<lightstep_project>.<%v_ID>'", resourceName, resourceName)
	}

	projectName, id := ids[0], ids[1]
	inferredServiceRuleResponse, err := apiClient.GetInferredServiceRule(ctx, projectName, id)
	if err != nil {
		return []*schema.ResourceData{}, fmt.Errorf("failed to get inferred service rule due to error: %v", err)
	}

	resourceData.SetId(id)
	if err = setResourceDataFromInferredServiceRule(projectName, &inferredServiceRuleResponse.Attributes, resourceData); err != nil {
		return nil, fmt.Errorf("failed to set inferred servie rule from API response to terraform state: %v", err)
	}

	return []*schema.ResourceData{resourceData}, nil
}

func getInferredServiceRuleAttributesFromResource(resourceData *schema.ResourceData) (client.InferredServiceRuleRequestAttributes, error) {
	requestAttributes := client.InferredServiceRuleRequestAttributes{
		Name:             getRuleNameFromResource(resourceData),
		Description:      getDescriptionFromResource(resourceData),
		AttributeFilters: getAttributeFiltersFromResource(resourceData),
		GroupByKeys:      getGroupByKeysFromResource(resourceData),
	}
	return requestAttributes, nil
}

func getProjectNameFromResource(resourceData *schema.ResourceData) string {
	return resourceData.Get("project_name").(string)
}

func getRuleNameFromResource(resourceData *schema.ResourceData) string {
	return resourceData.Get("name").(string)
}

func getDescriptionFromResource(resourceData *schema.ResourceData) *string {
	descriptionResource, ok := resourceData.GetOk("description")
	if !ok {
		return nil
	}
	description := descriptionResource.(string)
	return &description
}

func getAttributeFiltersFromResource(resourceData *schema.ResourceData) []client.AttributeFilter {
	attributeFiltersSet := resourceData.Get("attribute_filters").(*schema.Set)

	attributeFilters := make([]client.AttributeFilter, attributeFiltersSet.Len())

	for i, attributeFilterResource := range attributeFiltersSet.List() {
		attributeFilterAsMap := attributeFilterResource.(map[string]interface{})
		attributeFilters[i].Key = attributeFilterAsMap["key"].(string)

		valuesSet := attributeFilterAsMap["values"].(*schema.Set)
		for _, value := range valuesSet.List() {
			attributeFilters[i].Values = append(attributeFilters[i].Values, value.(string))
		}
	}

	return attributeFilters
}

func getGroupByKeysFromResource(resourceData *schema.ResourceData) []string {
	groupByKeysResource := resourceData.Get("group_by_keys")

	asInterfaceSlice := groupByKeysResource.([]interface{})
	stringSlice := make([]string, len(asInterfaceSlice))
	for i, element := range asInterfaceSlice {
		stringSlice[i] = element.(string)
	}
	return stringSlice
}

func setResourceDataFromInferredServiceRule(
	projectName string,
	responseAttributes *client.InferredServiceRuleResponseAttributes,
	resourceData *schema.ResourceData,
) error {
	if err := resourceData.Set("project_name", projectName); err != nil {
		return fmt.Errorf("unable to set project_name resource field: %v", err)
	}

	if err := resourceData.Set("name", responseAttributes.Name); err != nil {
		return fmt.Errorf("unable to set name resource field: %v", err)
	}

	if err := resourceData.Set("description", responseAttributes.Description); err != nil {
		return fmt.Errorf("unable to set description resource field: %v", err)
	}

	if err := resourceData.Set("attribute_filters", toAttributeFilterResources(responseAttributes.AttributeFilters)); err != nil {
		return fmt.Errorf("unable to set attribute_filters resource field: %v", err)
	}

	if err := resourceData.Set("group_by_keys", responseAttributes.GroupByKeys); err != nil {
		return fmt.Errorf("unable to set group_by_keys resource field: %v", err)
	}

	return nil
}

func toAttributeFilterResources(attributesFilters []client.AttributeFilter) []interface{} {
	attributeFilterResources := make([]interface{}, len(attributesFilters))
	for i, attributeFilter := range attributesFilters {
		attributeFilterResource := map[string]interface{}{}
		attributeFilterResource["key"] = attributeFilter.Key
		attributeFilterResource["values"] = attributeFilter.Values
		attributeFilterResources[i] = attributeFilterResource
	}

	return attributeFilterResources
}
