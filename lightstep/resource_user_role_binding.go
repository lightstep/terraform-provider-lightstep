package lightstep

import (
	"context"
	"fmt"
	"strings"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"

	"github.com/lightstep/terraform-provider-lightstep/client"
)

func resourceUserRoleBinding() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceUserRoleCreateOrUpdate,
		ReadContext:   resourceUserRoleBindingRead,
		UpdateContext: resourceUserRoleCreateOrUpdate,
		DeleteContext: resourceUserRoleBindingDelete,
		Importer: &schema.ResourceImporter{
			StateContext: resourceUserRoleBindingImport,
		},
		Schema: map[string]*schema.Schema{
			"role": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true, // changing role or project requires a new tf resource to ensure permissions are properly removed.
			},
			"project": {
				Type:     schema.TypeString,
				Optional: true,
				ForceNew: true, // changing role or project requires a new tf resource to ensure permissions are properly removed.
			},
			"users": {
				Type:     schema.TypeSet,
				Required: true,
				Elem: &schema.Schema{
					Type: schema.TypeString,
				},
			},
		},
	}
}

func getUserRoleBindingFromResource(_ context.Context, d *schema.ResourceData) client.RoleBinding {
	var userRoleBinding client.RoleBinding

	role, ok := d.GetOk("role")
	if ok {
		userRoleBinding.RoleName = role.(string)
	}

	project, ok := d.GetOk("project")
	if ok {
		userRoleBinding.ProjectName = project.(string)
	}

	users, ok := d.GetOk("users")
	if ok {
		for _, user := range users.(*schema.Set).List() {
			userRoleBinding.Users = append(userRoleBinding.Users, user.(string))
		}
	}

	return userRoleBinding
}

func setUserRoleBindingFromResource(d *schema.ResourceData, userRoleBinding client.RoleBinding) error {
	users := schema.NewSet(schema.HashString, nil)
	for _, user := range userRoleBinding.Users {
		users.Add(user)
	}

	err := d.Set("users", users)
	if err != nil {
		return fmt.Errorf("error to set users resource field: %v", err)
	}
	err = d.Set("role", userRoleBinding.RoleName)
	if err != nil {
		return fmt.Errorf("error to set role resource field: %v", err)
	}
	err = d.Set("project", userRoleBinding.ProjectName)
	if err != nil {
		return fmt.Errorf("error to set project resource field: %v", err)
	}

	return nil
}

func resourceUserRoleCreateOrUpdate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	c := m.(*client.Client)

	// Read the proposed plan
	userRoleBinding := getUserRoleBindingFromResource(ctx, d)

	// Update role binding
	_, err := c.UpdateRoleBinding(ctx, userRoleBinding.ProjectName, userRoleBinding.RoleName, userRoleBinding.Users...)
	if err != nil {
		return handleAPIError(err, d, "create/update user role binding")
	}

	// Save this resource with an ID with the following format "org_name:role_name:project_name",
	// where project name is optional.
	d.SetId(fmt.Sprintf("%s/%s", c.OrgName(), userRoleBinding.ID()))

	// update state by forcing a reading to the API.
	return resourceUserRoleBindingRead(ctx, d, m)
}

// resourceUserRoleBindingRead reads a user role binding from the resource data.
//
// When called by a Read or Delete Context, it will read data from the terraform state.
// When Called by a Create or Update context, it will reda data from the terraform plan.
func resourceUserRoleBindingRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	c := m.(*client.Client)

	// Read the resource from data.
	userRoleBinding := getUserRoleBindingFromResource(ctx, d)

	// Fetch role binding from the Lightstep API.
	userRoleBinding, err := c.ListRoleBinding(ctx, userRoleBinding.ProjectName, userRoleBinding.RoleName)
	if err != nil {
		return handleAPIError(err, d, "get user role binding")
	}

	// Store the fetched data in the state.
	err = setUserRoleBindingFromResource(d, userRoleBinding)
	if err != nil {
		return diag.FromErr(err)
	}

	return diags
}

func resourceUserRoleBindingDelete(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	c := m.(*client.Client)

	// Read from the terraform state
	userRoleBinding := getUserRoleBindingFromResource(ctx, d)

	// Update role binding with no users, this will remove this role from all users of this org for the given project.
	_, err := c.UpdateRoleBinding(ctx, userRoleBinding.ProjectName, userRoleBinding.RoleName)
	if err != nil {
		return handleAPIError(err, d, "delete user role binding")
	}

	// mark this resource as deleted.
	d.SetId("")

	return diag.Diagnostics{}
}

func resourceUserRoleBindingImport(ctx context.Context, d *schema.ResourceData, m interface{}) ([]*schema.ResourceData, error) {
	c := m.(*client.Client)

	ids := strings.Split(d.Id(), "/")
	if len(ids) < 2 {
		return nil, fmt.Errorf("user role binding id should be in the following format: {organization-name}/{role-name} or {organization-name}/{role-name}/{project-name}")
	}

	// User Role Binding ID has the following format: {organization-name}/{role-name}/{project-name}
	orgName := ids[0]
	roleName := ids[1]

	var projectName string
	if len(ids) == 3 {
		projectName = ids[2]
	}

	if c.OrgName() != orgName {
		return nil, fmt.Errorf("lightstep terraform provider is configured to organization %s but it is trying to import resource bindings from organization %s", c.OrgName(), orgName)
	}

	userRoleBinding, err := c.ListRoleBinding(ctx, projectName, roleName)
	if err != nil {
		return []*schema.ResourceData{}, fmt.Errorf("failed to get user role binding: %v", err)
	}

	d.SetId(d.Id())

	err = setUserRoleBindingFromResource(d, userRoleBinding)
	if err != nil {
		return nil, fmt.Errorf("error saving user role binding to state: %v", err)
	}

	return []*schema.ResourceData{d}, nil
}
