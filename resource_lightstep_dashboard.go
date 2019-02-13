package lightstep

import (
    "fmt"
    "github.com/hashicorp/terraform/helper/schema"
)

func resourceLightstepDashboard() *shema.Resource {
    return &schema.Resource {
        Create: resourceLightstepDashboardCreate,
    }
}

func resourceLightstepDashboardCreate(d *shema.ResourceData, meta interface{}) error {

}
