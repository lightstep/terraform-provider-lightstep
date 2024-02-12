package lightstep

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"regexp"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/hashicorp/terraform-plugin-sdk/v2/meta"

	"github.com/lightstep/terraform-provider-lightstep/client"
	"github.com/lightstep/terraform-provider-lightstep/version"
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
				DefaultFunc:  schema.EnvDefaultFunc("LIGHTSTEP_ENV", "public"),
				ValidateFunc: validation.StringMatch(regexp.MustCompile(`^(staging|meta|public)$`), "Must be one of: staging, meta, public"),
				Description:  "The name of the Lightstep environment, must be one of: staging, meta, public. Deprecated in favor of `api_url`",
				Deprecated:   "This field is deprecated and will be removed in a future release. Please use the `api_url` field instead.",
			},
			"api_url": {
				Type:        schema.TypeString,
				Optional:    true,
				DefaultFunc: schema.MultiEnvDefaultFunc([]string{"LIGHTSTEP_API_URL", "LIGHTSTEP_API_BASE_URL"}, ""),
				Description: "The base URL for the Lightstep API. This setting takes precedent over 'environment'. For example, https://api.lightstep.com",
			},
			"api_key": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: "The API Key for a Lightstep organization.",
			},
			// in order to support our internal use of lightstep in staging, meta, public
			// and avoid resetting LIGHTSTEP_API_KEY when switching environments
			// allow the user to specify where to look for the key
			"api_key_env_var": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: "Environment variable for Lightstep API key.",
				Default:     "LIGHTSTEP_API_KEY",
			},
		},

		ResourcesMap: map[string]*schema.Resource{
			"lightstep_stream":                 resourceStream(),
			"lightstep_stream_dashboard":       resourceStreamDashboard(),
			"lightstep_stream_condition":       resourceStreamCondition(),
			"lightstep_metric_condition":       resourceUnifiedCondition(MetricConditionSchema),
			"lightstep_metric_dashboard":       resourceUnifiedDashboard(MetricChartSchema),
			"lightstep_webhook_destination":    resourceWebhookDestination(),
			"lightstep_pagerduty_destination":  resourcePagerdutyDestination(),
			"lightstep_slack_destination":      resourceSlackDestination(),
			"lightstep_servicenow_destination": resourceServiceNowDestination(),
			"lightstep_alerting_rule":          resourceAlertingRule(),
			"lightstep_dashboard":              resourceUnifiedDashboard(UnifiedChartSchema),
			"lightstep_alert":                  resourceUnifiedCondition(UnifiedConditionSchema),
			"lightstep_user_role_binding":      resourceUserRoleBinding(),
			"lightstep_inferred_service_rule":  resourceInferredServiceRule(),
			"lightstep_saml_group_mappings":    resourceSAMLGroupMappings(),
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
	var apiKey string
	apiKey = d.Get("api_key").(string)
	if len(apiKey) == 0 {
		envVar := d.Get("api_key_env_var").(string)
		apiKeyEnv, ok := os.LookupEnv(envVar)
		if !ok {
			diags = append(diags, diag.Diagnostic{
				Severity: diag.Error,
				Summary:  "No api key found",
				Detail:   fmt.Sprintf("'api_key_env_var' is set to %v - but no api key found.", envVar),
			})
			return apiKey, diags
		}
		apiKey = apiKeyEnv
	}

	// TODO remove this code once `environment` is fully deprecated
	var baseUrl string
	optionalUrl := d.Get("api_url").(string)
	if len(optionalUrl) > 0 {
		baseUrl = optionalUrl
	} else {
		// get the url from the `environment` provider attribute
		env := d.Get("environment").(string)
		if env == "public" {
			baseUrl = "https://api.lightstep.com"
		} else {
			baseUrl = fmt.Sprintf("https://api-%v.lightstep.com", env)
		}
	}

	client := client.NewClientWithUserAgent(
		apiKey,
		d.Get("organization").(string),
		baseUrl,
		fmt.Sprintf("%s/%s (terraform %s)", "terraform-provider-lightstep", version.ProviderVersion, meta.SDKVersionString()),
	)

	return client, diags
}

func handleAPIError(err error, d *schema.ResourceData, resourceName string) diag.Diagnostics {
	apiErr, ok := err.(client.APIResponseCarrier)
	if ok {
		if apiErr.GetStatusCode() == http.StatusNotFound {
			d.SetId("")
			return diag.FromErr(fmt.Errorf("%s not found: %v", resourceName, apiErr))
		}
	}

	return diag.FromErr(fmt.Errorf("failed to %s: %v", resourceName, err))
}
