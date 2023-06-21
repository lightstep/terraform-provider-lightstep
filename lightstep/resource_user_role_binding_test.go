package lightstep

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

func TestAccUserRoleBinding(t *testing.T) {
	validConfiguration := `
resource "lightstep_user_role_binding" "org_admin" {
	role = "Organization Editor"
	users = [
		"terraform-test+1@lightstep.com",
		"terraform-test+2@lightstep.com"
	]
}

resource "lightstep_user_role_binding" "org_restricted" {
	role = "Organization Restricted Member"
	users = [
		"terraform-test+3@lightstep.com",
		"terraform-test+4@lightstep.com",
		"terraform-test+5@lightstep.com"
	]
}

resource "lightstep_user_role_binding" "proj_editor" {
	project = "` + testProject + `"
	role = "Project Editor"
	users = [
		"terraform-test+3@lightstep.com",
		"terraform-test+4@lightstep.com"
	]
}

resource "lightstep_user_role_binding" "proj_member" {
	project = "` + testProject + `"
	role = "Project Member"
	users = [
		"terraform-test+5@lightstep.com"
	]
}
`

	updatedConfiguration := `
resource "lightstep_user_role_binding" "org_admin" {
	role = "Organization Editor"
	users = [
		"terraform-test+2@lightstep.com"
	]
}

resource "lightstep_user_role_binding" "org_restricted" {
	role = "Organization Restricted Member"
	users = [
		"terraform-test+1@lightstep.com",
		"terraform-test+3@lightstep.com",
		"terraform-test+4@lightstep.com",
		"terraform-test+5@lightstep.com"
	]
}

resource "lightstep_user_role_binding" "proj_editor" {
	project = "` + testProject + `"
	role = "Project Editor"
	users = [
		"terraform-test+1@lightstep.com",
		"terraform-test+3@lightstep.com",
		"terraform-test+4@lightstep.com"
	]
}


resource "lightstep_user_role_binding" "proj_member" {
	project = "` + testProject + `"
	role = "Project Member"
	users = [
	]
}
`

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: testAccProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: validConfiguration,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("lightstep_user_role_binding.org_admin", "project", ""),
					resource.TestCheckResourceAttr("lightstep_user_role_binding.org_admin", "users.#", "2"),
					resource.TestCheckTypeSetElemAttr("lightstep_user_role_binding.org_admin", "users.*", "terraform-test+1@lightstep.com"),
					resource.TestCheckTypeSetElemAttr("lightstep_user_role_binding.org_admin", "users.*", "terraform-test+2@lightstep.com"),

					resource.TestCheckResourceAttr("lightstep_user_role_binding.org_restricted", "project", ""),
					resource.TestCheckResourceAttr("lightstep_user_role_binding.org_restricted", "users.#", "3"),
					resource.TestCheckTypeSetElemAttr("lightstep_user_role_binding.org_restricted", "users.*", "terraform-test+3@lightstep.com"),
					resource.TestCheckTypeSetElemAttr("lightstep_user_role_binding.org_restricted", "users.*", "terraform-test+4@lightstep.com"),
					resource.TestCheckTypeSetElemAttr("lightstep_user_role_binding.org_restricted", "users.*", "terraform-test+5@lightstep.com"),

					resource.TestCheckResourceAttr("lightstep_user_role_binding.proj_editor", "project", testProject),
					resource.TestCheckResourceAttr("lightstep_user_role_binding.proj_editor", "users.#", "2"),
					resource.TestCheckTypeSetElemAttr("lightstep_user_role_binding.proj_editor", "users.*", "terraform-test+3@lightstep.com"),
					resource.TestCheckTypeSetElemAttr("lightstep_user_role_binding.proj_editor", "users.*", "terraform-test+4@lightstep.com"),

					resource.TestCheckResourceAttr("lightstep_user_role_binding.proj_member", "project", testProject),
					resource.TestCheckResourceAttr("lightstep_user_role_binding.proj_member", "users.#", "1"),
					resource.TestCheckTypeSetElemAttr("lightstep_user_role_binding.proj_member", "users.*", "terraform-test+5@lightstep.com"),
				),
			},
			{
				Config: updatedConfiguration,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("lightstep_user_role_binding.org_admin", "project", ""),
					resource.TestCheckResourceAttr("lightstep_user_role_binding.org_admin", "users.#", "1"),
					resource.TestCheckTypeSetElemAttr("lightstep_user_role_binding.org_admin", "users.*", "terraform-test+2@lightstep.com"),

					resource.TestCheckResourceAttr("lightstep_user_role_binding.org_restricted", "project", ""),
					resource.TestCheckResourceAttr("lightstep_user_role_binding.org_restricted", "users.#", "4	"),
					resource.TestCheckTypeSetElemAttr("lightstep_user_role_binding.org_restricted", "users.*", "terraform-test+1@lightstep.com"),
					resource.TestCheckTypeSetElemAttr("lightstep_user_role_binding.org_restricted", "users.*", "terraform-test+3@lightstep.com"),
					resource.TestCheckTypeSetElemAttr("lightstep_user_role_binding.org_restricted", "users.*", "terraform-test+4@lightstep.com"),
					resource.TestCheckTypeSetElemAttr("lightstep_user_role_binding.org_restricted", "users.*", "terraform-test+5@lightstep.com"),

					resource.TestCheckResourceAttr("lightstep_user_role_binding.proj_editor", "project", testProject),
					resource.TestCheckResourceAttr("lightstep_user_role_binding.proj_editor", "users.#", "3"),
					resource.TestCheckTypeSetElemAttr("lightstep_user_role_binding.proj_editor", "users.*", "terraform-test+1@lightstep.com"),
					resource.TestCheckTypeSetElemAttr("lightstep_user_role_binding.proj_editor", "users.*", "terraform-test+3@lightstep.com"),
					resource.TestCheckTypeSetElemAttr("lightstep_user_role_binding.proj_editor", "users.*", "terraform-test+4@lightstep.com"),

					resource.TestCheckResourceAttr("lightstep_user_role_binding.proj_member", "project", testProject),
					resource.TestCheckResourceAttr("lightstep_user_role_binding.proj_member", "users.#", "0"),
				),
			},
		},
	})

}
