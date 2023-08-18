# The following example configures two users with the "Organization Restricted Member" role.
# It also overrides their permissions in "Project A", one with the "Project Editor" role and the other with the "Project Viewer" role.

# Set users to "Organization Restricted Member"
resource "lightstep_user_role_binding" "org_restricted" {
  role = "Organization Restricted Member"
  users = [
    "proj_a-editor@lightstep.com",
    "proj_a-viewer@lightstep.com"
  ]
}

# Set user "proj_a-editor@lightstep.com" to Project Editor in "Project A"
resource "lightstep_user_role_binding" "proj_editor" {
  project = "Project A"
  role    = "Project Editor"
  users = [
    "proj_a-editor@lightstep.com",
  ]
}

# Set user "proj_a-viewer@lightstep.com" to Project Viewer in "Project A"
resource "lightstep_user_role_binding" "proj_viewer" {
  project = "Project A"
  role    = "Project Viewer"
  users = [
    "proj_a-viewer@lightstep.com"
  ]
}
