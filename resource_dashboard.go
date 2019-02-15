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
                Optional: true,
            },
            "name": {
                Type: schema.TypeString,
                Optional: true,
            },
            "project": {
                Type: schema.TypeString,
                Required: true,
            },
            "searchAttributes": {
            resource_lightstep_dashboard.go
                Type: schema.TypeList,
                Optional: true,
                Elem: &schema.Resource{
                    Schema: map[string]*schema.Schema{
                        "name": &schema.Schema{
                            Type: schema.TypeString,
                            Required: true,
                        },
                        "query": &schema.Schema{
                            Type: schema.TypeString,
                            Required: true,
                        },
                    },
                },
            },
        },
    }
}

func resourceDashboardCreate(d *schema.ResourceData, meta interface{}) error {
    client := meta.(*lightstep.Client)

    var searchAttributes []lightstep.SearchAttributes
    for _, sa := range d.Get("searchAttributes").([]interface{}) {
        searchAttributes = append(
                searchAttributes,
                lightstep.SearchAttributes{
                    Name: sa.Get("name").(string),
                    Query: sa.Get("query").(string),
                },)
    }


    _, err := client.CreateDashboard(
        d.Get("project").(string),
        d.Get("name").(string),
        searchAttributes,
    )
    if err != nil {
        return err
    }

    return resourceLightstepDashboardRead(d, meta)
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

    return nil
}

func resourceDashboardDelete(d *schema.ResourceData, meta interface{}) error {
    client := meta.(*lightstep.Client)
    client.DeleteDashboard(
        d.Get("project").(string),
        d.Get("dashboard_id").(string),
    )

    return nil
}

func resourceLightstepDashboardUpdate(d *schema.ResourceData, meta interface{}) error {
    return resourceLightstepDashboardRead(d, meta);
}
