package lightstep

import (
    "fmt"
    "github.com/hashicorp/terraform/helper/schema"
    lightstep "github.com/lightstep/terraform-provider-lightstep/lightstep"
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

func resourceLightstepDashboardCreate(d *schema.ResourceData, meta interface{}) error {
    client := meta.(*lightstep.Client)

}

func resourceLightstepDashboardGet(d *schema.ResourceData, meta interface{}) error {
    client := meta.(*lightstep.Client)


}

func resourceLightstepDashboardDelete(d *schema.ResourceData, meta interface{}) {
    client := meta.(*lightstep.Client)


}
