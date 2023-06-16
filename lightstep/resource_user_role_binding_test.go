package lightstep

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

func TestAccUserRoleBinding(t *testing.T) {
	t.Skip()

	validConfiguration := `
resource "lightstep_user_role_binding" "org_admin" {
	role_name = "Organization Editor"
	users = [
		"terraform-test+1@lightstep.com",
		"terraform-test+2@lightstep.com"
	]
}

resource "lightstep_user_role_binding" "org_restricted" {
	role_name = "Organization Restricted Member"
	users = [
		"terraform-test+3@lightstep.com",
		"terraform-test+4@lightstep.com"
	]
}

resource "lightstep_user_role_binding" "proj_editor" {
	project_name = "` + testProject + `"
	role_name = "Project Editor"
	users = [
		"terraform-test+3@lightstep.com",
		"terraform-test+4@lightstep.com"
	]
}
`

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		Providers:         testAccProviders,
		ProviderFactories: testAccProviderFactories,
		CheckDestroy:      testAccPagerdutyDestinationDestroy,
		Steps: []resource.TestStep{
			{
				Config: validConfiguration,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("lightstep_user_role_binding.org_admin", "users.#", "2"),
					resource.TestCheckResourceAttr("lightstep_user_role_binding.org_admin", "users.0", "terraform-test+1@lightstep.com"),
					resource.TestCheckResourceAttr("lightstep_user_role_binding.org_admin", "users.1", "terraform-test+2@lightstep.com"),

					resource.TestCheckResourceAttr("lightstep_user_role_binding.proj_editor", "users.#", "2"),
					resource.TestCheckResourceAttr("lightstep_user_role_binding.proj_editor", "users.0", "terraform-test+1@lightstep.com"),
					resource.TestCheckResourceAttr("lightstep_user_role_binding.proj_editor", "users.1", "terraform-test+2@lightstep.com"),

					resource.TestCheckResourceAttr("lightstep_user_role_binding.org_restricted", "users.#", "2"),
					resource.TestCheckResourceAttr("lightstep_user_role_binding.org_restricted", "users.0", "terraform-test+1@lightstep.com"),
					resource.TestCheckResourceAttr("lightstep_user_role_binding.org_restricted", "users.1", "terraform-test+2@lightstep.com"),
				),
			},
		},
	})

}
