package main

import (
    "github.com/hashicorp/terraform/helper/schema"
    "github.com/lightstep/terraform-provider-lightstep/lightstep"
)

func resourceLightstepDashboard() *schema.Resource {
    return &schema.Resource {
        Create: resourceLightstepDashboardCreate,
        Read:    resourceLightstepDashboardRead,
        Delete: resourceLightstepDashboardDelete,

        Schema: map[string]*schema.Schema {
            "dashboard_id": &schema.Schema {
                Type: schema.TypeString,
                Required: false,
            },
            "name": &schema.Schema {
                Type: schema.TypeString,
                Required: false,
            },
            "project": &schema.Schema {
                Type: schema.TypeString,
                Required: true,
            },
            "searchAttributes": &schema.Schema {
                Type: schema.TypeList,
                Required: false,
            },
        },
    }
}

func resourceLightstepDashboardCreate(d *schema.ResourceData, meta interface{}) error {
    client := meta.(*lightstep.Client)
    _, err := client.CreateDashboard(
        d.Get("project").(string),
        d.Get("name").(string),
        d.Get("searchAttributes").([]lightstep.SearchAttributes),
    )
    if err != nil {
        return err
    }

    return resourceStreamRead(d, meta)
}

func resourceLightstepDashboardRead(d *schema.ResourceData, meta interface{}) error {
    client := meta.(*lightstep.Client)
    _, err := client.GetDashboard(
        d.Get("project").(string),
        d.Get("dashboard_id").(string),
    )

    if err != nil {
        return err
    }

    return resourceStreamRead(d, meta)
}

func resourceLightstepDashboardDelete(d *schema.ResourceData, meta interface{}) error {
    client := meta.(*lightstep.Client)
    client.DeleteDashboard(
        d.Get("project").(string),
        d.Get("dashboard_id").(string),
    )

    return nil
}
