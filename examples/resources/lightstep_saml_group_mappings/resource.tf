resource "lightstep_saml_group_mappings" "group_mappings" {
  mapping {
    # For users with the "member_of: sre" SAML Attribute,
    # Assign "Organization Editor" in the organization.
    match {
      attribute_key   = "member_of"
      attribute_value = "sre"
    }
    roles {
      organization_role = "Organization Editor"
    }
  }

  mapping {
    # For users with the "member_of: developer" SAML Attribute,
    # Assign "Organization Restricted Member" in the organization, and
    # Assign "Project Viewer" in the "Project A" project.

    match {
      attribute_key   = "member_of"
      attribute_value = "developer"
    }
    roles {
      organization_role = "Organization Restricted Member"
      project_roles = {
        "Project A" = "Project Viewer"
      }
    }
  }
}