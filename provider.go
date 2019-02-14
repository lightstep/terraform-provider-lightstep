package main

import (
	"github.com/hashicorp/terraform/helper/schema"
)

func Provider() *schema.Provider {
	return &schema.Provider{
                Schema: map[string]*schema.Schema{
                        "api_key": {
                                Type: schema.TypeString,
                                Required: true,
                                DefaultFunc: schema.EnvDefaultFunc("LIGHTSTEP_API_KEY", nil),
                        },
                        "base_url": {
                                Type: schema.TypeString,
                                Required: false,
                                DefaultFunc: schema.EnvDefaultFunc("LIGHTSTEP_BASE_URL", "https://api-staging.lightstep.com/public/v0.1/"),
                        },
                        "org_name": {
                                Type: schema.TypeString,
                                Required: true,
                                DefaultFunc: schema.EnvDefaultFunc("LIGHTSTEP_ORG_NAME", nil),
                        }
                },
		ResourcesMap: map[string]*schema.Resource{
			"lightstep_stream": resourceStream(),
                        "lightstep_dashboard": resourceLightstepDashboard(),
		},
	}
}
