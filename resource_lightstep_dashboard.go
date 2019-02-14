package main

import (
    "github.com/hashicorp/terraform/helper/schema"
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
                Required: true,
            },
            "projectName": &schema.Schema {
                Type: schema.TypeString,
                Required: true,
            },
            "searchAttributes": &schema.Schema {
                Type: schema.TypeString,
                Required: false,
            },
        },
    }
}

func resourceLightstepDashboardCreate(d *schema.ResourceData, meta interface{}) error {
    return nil
}

func resourceLightstepDashboardRead(d *schema.ResourceData, meta interface{}) error {
    return nil
}

func resourceLightstepDashboardDelete(d *schema.ResourceData, meta interface{}) error {
    return nil
}
