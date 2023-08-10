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

func resourceUserRoleBinding() *schema.Resource {
	return &schema.Resource{
		Description: `Provides a [Lightstep Role Binding](https://api-docs.lightstep.com/reference/RoleBinding). This can be used to manage User's Organization level roles and Project level roles.

A role binding can target either the Organization level roles or a Project role for a specific project. An user Project role can't be set to a more restrict role than their Organization level role. 

**NOTE**: this terraform resource is authoritative, users that are not declared in a terraform resource will lose the declared role in the specified organization/project.

The list of valid roles for Organization level role bindings are:
* Organization Admin
* Organization Editor
* Organization Viewer


The list of valid roles for Project level role bindings are:
* Project Editor
* Project Viewer


Changes to both Organization level role and Project level roles for the same user can cause race condition, 
in that case we suggest these changes to be made in two steps. 
* When lowering an user Organization level role and upping their Project level Role, first change their organization role.
* When upping an user Organization level role and removing or lowering their Project level Role, first change their project role.
`,
		CreateContext: resourceUserRoleBindingCreateOrUpdate,
		ReadContext:   resourceUserRoleBindingRead,
		UpdateContext: resourceUserRoleBindingCreateOrUpdate,
		DeleteContext: resourceUserRoleBindingDelete,
		Importer: &schema.ResourceImporter{
			StateContext: resourceUserRoleBindingImport,
		},
		Schema: map[string]*schema.Schema{
			"role": {
				Type:        schema.TypeString,
				Required:    true,
				ForceNew:    true, // changing role or project requires a new tf resource to ensure permissions are properly removed.
				Description: "Role's name being granted with this role binding.",
				ValidateFunc: validation.StringInSlice([]string{
					"Organization Admin",
					"Organization Editor",
					"Organization Viewer",
					"Organization Restricted Member",
					"Project Editor",
					"Project Viewer",
				}, false),
			},
			"project": {
				Type:        schema.TypeString,
				Optional:    true,
				ForceNew:    true, // changing role or project requires a new tf resource to ensure permissions are properly removed.
				Description: "Name of the project where this role will be applied; if omitted the role will be applied to the organization",
			},
			"users": {
				Type:     schema.TypeSet,
				Required: true,
				Elem: &schema.Schema{
					Type: schema.TypeString,
				},
				Description: "Complete list of users that should have this specified role in the organization or in the project (if specified). Important: this list is authoritative; any users not included in this list WILL NOT have this role for the given project or organization.",
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

func resourceUserRoleBindingCreateOrUpdate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
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
// When called by a Create or Update context, it will read data from the terraform plan.
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
		return nil, fmt.Errorf("lightstep terraform provider is configured to organization %s but it is trying to import resource bindings from organization %s: %s", c.OrgName(), orgName, d.Id())
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
