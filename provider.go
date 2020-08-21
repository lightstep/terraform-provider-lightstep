package main

import (
	"context"
	"regexp"

	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/helper/validation"
	"github.com/lightstep/terraform-provider-lightstep/lightstep"
)

func Provider() *schema.Provider {
	return &schema.Provider{
		Schema: map[string]*schema.Schema{
			"api_key": {
				Type:        schema.TypeString,
				Required:    true,
				DefaultFunc: schema.EnvDefaultFunc("LIGHTSTEP_API_KEY", nil),
				Description: "A Lightstep API Key with Member role. To create one follow the instructions here:https://docs.lightstep.com/docs/create-and-manage-api-keys",
			},
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
	client := lightstep.NewClient(
		context.Background(),
		d.Get("api_key").(string),
		d.Get("organization").(string),
		d.Get("environment").(string),
	)

	return client, nil
}
