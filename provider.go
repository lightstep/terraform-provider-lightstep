package main

import (
  "context"
  "log"
	"github.com/hashicorp/terraform/helper/schema"
  "github.com/lightstep/terraform-provider-lightstep/lightstep"
)

func Provider() *schema.Provider {
	return &schema.Provider{
    Schema: map[string]*schema.Schema{
      "api_key": {
        Type:        schema.TypeString,
        Required:    true,
        DefaultFunc: schema.EnvDefaultFunc("LIGHTSTEP_API_KEY", nil),
      },
      "organization": {
        Type:        schema.TypeString,
        Required:    true,
        DefaultFunc: schema.EnvDefaultFunc("LIGHTSTEP_HOST", nil),
      },
    },

		ResourcesMap: map[string]*schema.Resource{
			"lightstep_project": resourceProject(),
      // "lightstep_dashboard": resourceDashboard(),
      "lightstep_stream": resourceStream(),
		},

    ConfigureFunc: configureProvider,
	}
}

func configureProvider(d *schema.ResourceData) (interface{}, error) {
  // Initializing client here with credentials
  client := lightstep.NewClient(
    context.Background(),
    d.Get("api_key").(string),
    d.Get("organization").(string),
  )
  log.Println("[INFO] LightStep client successfully initialized.")
  // panic("here")

  return client, nil
}
