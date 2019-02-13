package lightstep

import (
    "fmt"
    "github.com/hashicorp/terraform/helper/schema"
)

func resourceLightstepDashboard() *shema.Resource {
    return &schema.Resource {
        Create: resourceLightstepDashboardCreate,
        Get:    resourceLightstepDashboardGet,
        Delete: resourceLightstepDashboardDelete,
    },

    Schema: map[string]*schema.Schema {

    }
}

func resourceLightstepDashboardCreate(d *shema.ResourceData, meta interface{}) error {
    client := meta.(*lightstep.Client)

}

func resourceLightstepDashboardGet() {
    client := meta.(*lightstep.Client)
}

func resourceLightstepDashboardDelete() {
    client := meta.(*lightstep.Client)
}
