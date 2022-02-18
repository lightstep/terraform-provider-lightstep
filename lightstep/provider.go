package lightstep

import (
	"context"
	"fmt"
	"os"
	"regexp"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/hashicorp/terraform-plugin-sdk/v2/meta"
	"github.com/lightstep/terraform-provider-lightstep/client"
	"github.com/lightstep/terraform-provider-lightstep/version"
)

// Provider returns the build lightstep provider object
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
				DefaultFunc:  schema.EnvDefaultFunc("LIGHTSTEP_ENV", "public"),
				ValidateFunc: validation.StringMatch(regexp.MustCompile(`^(staging|meta|public)$`), "Must be one of: staging, meta, public"),
				Description:  "The name of the Lightstep environment, must be one of: staging, meta, public.",
			},
			// in order to support our internal use of lightstep in staging, meta, public
			// and avoid resetting LIGHTSTEP_API_KEY when switching environments
			// allow the user to specify where to look for the key
			"api_key_env_var": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: "Environment variable for Lightstep api key. Will only be used if `api_key` is not set",
				Default:     "LIGHTSTEP_API_KEY",
			},
			"api_key": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: "Lightstep API key. Will override `api_key_env_var` if set.",
				Default:     "",
			},
		},

		ResourcesMap: map[string]*schema.Resource{
			"lightstep_stream":                resourceStream(),
			"lightstep_stream_dashboard":      resourceStreamDashboard(),
			"lightstep_stream_condition":      resourceStreamCondition(),
			"lightstep_metric_condition":      resourceMetricCondition(),
			"lightstep_metric_dashboard":      resourceMetricDashboard(),
			"lightstep_webhook_destination":   resourceWebhookDestination(),
			"lightstep_pagerduty_destination": resourcePagerdutyDestination(),
			"lightstep_slack_destination":     resourceSlackDestination(),
			"lightstep_alerting_rule":         resourceAlertingRule(),
		},

		DataSourcesMap: map[string]*schema.Resource{
			"lightstep_stream": dataSourceStream(),
		},

		ConfigureContextFunc: configureProvider,
		TerraformVersion:     "v1.0.3",
	}
}

func configureProvider(_ context.Context, d *schema.ResourceData) (interface{}, diag.Diagnostics) {
	var diags diag.Diagnostics
	apiKey := d.Get("api_key_env_var").(string)

	if len(apiKey) == 0 {
		envVar := d.Get("api_key_env_var").(string)

		var ok bool
		apiKey, ok = os.LookupEnv(envVar)
		if !ok {
			diags = append(diags, diag.Diagnostic{
				Severity: diag.Error,
				Summary:  "No api key found",
				Detail:   fmt.Sprintf("'api_key_env_var' is set to %v - but no api key found.", envVar),
			})
			return apiKey, diags
		}
	}

	client := client.NewClientWithUserAgent(
		apiKey,
		d.Get("organization").(string),
		d.Get("environment").(string),
		fmt.Sprintf("%s/%s (terraform %s)", "terraform-provider-lightstep", version.ProviderVersion, meta.SDKVersionString()),
	)

	return client, diags
}
