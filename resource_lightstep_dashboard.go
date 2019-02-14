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
        "dashboard_id": &schema.Schema {
            type: schema.TypeString,
            required: false
        },
        "name": &schema.Schema {
            type: schema.TypeString,
            required: true
        },
        "projectName": &schema.Schema {
            type: schema.TypeString,
            required: true
        },
        "searchAttributes": &schema.Schema {
            type: schema.TypeArray,
            required: false
        }
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
