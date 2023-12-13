package lightstep

import (
	"regexp"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

func TestAccSAMLGroupMappings(t *testing.T) {
	validMappings := `
resource "lightstep_saml_group_mappings" "group_mappings" {
  mapping {
    match {
      attribute_key = "member_of"
      attribute_value = "sre"
    }
    roles {
      organization_role = "Organization Editor"
    }
  }

  mapping {
    match {
      attribute_key = "member_of"
      attribute_value = "developer"
    }
    roles {
      organization_role = "Organization Restricted Member"
      project_roles = {
            ` + testProject + ` =  "Project Viewer"
      }
    }
  }
}
`

	missingOrgRole := `
resource "lightstep_saml_group_mappings" "group_mappings" {
  mapping {
    match {
      attribute_key = "member_of"
      attribute_value = "sre"
    }
    roles {
      organization_role = "Organization Editor"
    }
  }

  mapping {
    match {
      attribute_key = "member_of"
      attribute_value = "developer"
    }
    roles {
      project_roles = {
            ` + testProject + ` =  "Project Viewer"
      }
    }
  }
}
`

	updatedMappings := `
resource "lightstep_saml_group_mappings" "group_mappings" {
  mapping {
    match {
      attribute_key = "member_of"
      attribute_value = "frontend"
    }
    roles {
      organization_role = "Organization Editor"
    }
  }

  mapping {
    match {
      attribute_key = "member_of"
      attribute_value = "backend"
    }
    roles {
      organization_role = "Organization Restricted Member"
      project_roles = {
            ` + testProject + ` =  "Project Viewer"
      }
    }
  }
}
`

	singleMapping := `
resource "lightstep_saml_group_mappings" "group_mappings" {
  mapping {
    match {
      attribute_key = "member_of"
      attribute_value = "frontend"
    }
    roles {
      organization_role = "Organization Editor"
    }
  }
}
`

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccPagerdutyDestinationDestroy,
		Steps: []resource.TestStep{
			{
				Config: validMappings,
			},
			{
				Config:      missingOrgRole,
				ExpectError: regexp.MustCompile("The argument \"organization_role\" is required, but no definition was found."),
			},
			{
				Config: updatedMappings,
			},
			{
				Config: singleMapping,
			},
		},
	})

}
