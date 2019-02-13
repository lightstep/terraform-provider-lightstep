package main

import (
	"github.com/hashicorp/terraform/helper/schema"
)

func Provider() *schema.Provider {
	return &schema.Provider{
		ResourcesMap: map[string]*schema.Resource{
			"lightstep_stream": resourceStream(),
		},

                ResourcesMap: map[string]*schema.Resource{
                    "lightstep_dashboard": resourceLightstepDashboard(),
                },
	}
}
