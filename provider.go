package main

import (
	"context"
	"fmt"
	"os"
	"regexp"

	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/helper/validation"
	"github.com/lightstep/terraform-provider-lightstep/lightstep"
)

func Provider() *schema.Provider {
	return &schema.Provider{
		Schema: map[string]*schema.Schema{
			"organization": {
				Type:        schema.TypeString,
				Required:    true,
				DefaultFunc: schema.EnvDefaultFunc("LIGHTSTEP_ORG", nil),
				Description: "The name of the Lightstep organization",
			},
			"environment": {
				Type:         schema.TypeString,
				Optional:     true,
				Default:      "public",
				ValidateFunc: validation.StringMatch(regexp.MustCompile(`^(staging|meta|public)$`), "Must be one of: staging, meta, public"),
				Description:  "The name of the Lightstep environment, must be one of: staging, meta, public.",
			},
			// in order to support our internal use of lightstep in staging, meta, public
			// and avoid resetting LIGHTSTEP_API_KEY when switching environments
			// allow the user to specify where to look for the key
			"api_key_env_var": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: "Environment variable for Lightstep api key.",
				Default:     "LIGHTSTEP_API_KEY",
			},
		},

		ResourcesMap: map[string]*schema.Resource{
			"lightstep_dashboard":             resourceDashboard(),
			"lightstep_stream":                resourceStream(),
			"lightstep_condition":             resourceCondition(),
			"lightstep_webhook_destination":   resourceWebhookDestination(),
			"lightstep_pagerduty_destination": resourcePagerdutyDestination(),
		},

		ConfigureFunc:    configureProvider,
		TerraformVersion: "v.12.26",
	}
}

func configureProvider(d *schema.ResourceData) (interface{}, error) {
	envVar := d.Get("api_key_env_var").(string)

	apiKey, ok := os.LookupEnv(envVar)
	if !ok {
		return apiKey, fmt.Errorf("'api_key_env_var' is set to %v - but no api key found.", envVar)
	}

	client := lightstep.NewClient(
		context.Background(),
		apiKey,
		d.Get("organization").(string),
		d.Get("environment").(string),
	)

	return client, nil
}
