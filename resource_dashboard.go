package main

import (
    "github.com/hashicorp/terraform/helper/schema"
    "github.com/lightstep/terraform-provider-lightstep/lightstep"
)

func resourceDashboard() *schema.Resource {
    return &schema.Resource {
        Create: resourceDashboardCreate,
        Read:    resourceDashboardRead,
        Delete: resourceDashboardDelete,

        Schema: map[string]*schema.Schema {
            "dashboard_id": {
                Type: schema.TypeString,
                Required: false,
            },
            "name": {
                Type: schema.TypeString,
                Required: false,
            },
            "project": {
                Type: schema.TypeString,
                Required: true,
            },
            "searchAttributes": {
                Type: schema.TypeList,
                Required: false,
            },
        },
    }
}

func resourceDashboardCreate(d *schema.ResourceData, meta interface{}) error {
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

func resourceDashboardRead(d *schema.ResourceData, meta interface{}) error {
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

func resourceDashboardDelete(d *schema.ResourceData, meta interface{}) error {
    client := meta.(*lightstep.Client)
    client.DeleteDashboard(
        d.Get("project").(string),
        d.Get("dashboard_id").(string),
    )

    return nil
}
