package lightstep

import (
	"context"
	"fmt"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/provider"
	"github.com/hashicorp/terraform-plugin-framework/provider/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-sdk/v2/meta"
	"github.com/lightstep/terraform-provider-lightstep/version"
	"os"
	"regexp"

	"github.com/lightstep/terraform-provider-lightstep/client"
)

type lightstepProvider struct {
	*client.Client
}

type LightstepProviderModel struct {
	Organization types.String `tfsdk:"organization"`
	Environment  types.String `tfsdk:"environment"`
	ApiUrl       types.String `tfsdk:"api_url"`
	ApiKey       types.String `tfsdk:"api_key"`
	ApiKeyEnvVar types.String `tfsdk:"api_key_env_var"`
}

func New() provider.Provider {
	return &lightstepProvider{}
}

func (p *lightstepProvider) Metadata(ctx context.Context, req provider.MetadataRequest, resp *provider.MetadataResponse) {
	resp.TypeName = "lightstep"
}

func (p *lightstepProvider) Schema(ctx context.Context, req provider.SchemaRequest, resp *provider.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"organization": schema.StringAttribute{
				Optional:    true,
				Description: "The name of the Lightstep organization",
			},
			"environment": schema.StringAttribute{
				Optional:           true,
				Description:        "The name of the Lightstep environment, must be one of: staging, meta, public. Deprecated in favor of `api_url`",
				DeprecationMessage: "This field is deprecated and will be removed in a future release. Please use the `api_url` field instead.",
			},
			"api_url": schema.StringAttribute{
				Optional:    true,
				Description: "The base URL for the Lightstep API. This setting takes precedent over 'environment'. For example, https://api.lightstep.com",
			},
			"api_key": schema.StringAttribute{
				Optional:    true,
				Description: "The API Key for a Lightstep organization.",
			},
			// in order to support our internal use of lightstep in staging, meta, public
			// and avoid resetting LIGHTSTEP_API_KEY when switching environments
			// allow the user to specify where to look for the key
			"api_key_env_var": schema.StringAttribute{
				Optional:    true,
				Description: "Environment variable for Lightstep API key.",
			},
		},
	}
}

func (p *lightstepProvider) Configure(ctx context.Context, req provider.ConfigureRequest, resp *provider.ConfigureResponse) {
	var data LightstepProviderModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)

	// apply defaults
	if len(data.Organization.ValueString()) == 0 {
		if orgFromEnv, hasOrgFromEnv := os.LookupEnv("LIGHTSTEP_ORG"); hasOrgFromEnv {
			data.Organization = types.StringValue(orgFromEnv)
		} else {
			resp.Diagnostics.AddError("organization is not set", "set organization in your HCL or set the LIGHTSTEP_ORG env var")
			return
		}
	}

	if len(data.Environment.ValueString()) == 0 {
		if envFromEnv, hasEnvFromEnv := os.LookupEnv("LIGHTSTEP_ENV"); hasEnvFromEnv {
			data.Environment = types.StringValue(envFromEnv)
		} else {
			data.Environment = types.StringValue("public")
		}
	}
	if !regexp.MustCompile(`^(staging|meta|public)$`).MatchString(data.Environment.ValueString()) {
		resp.Diagnostics.AddError(fmt.Sprintf("environment %s is invalid", data.Environment.String()), "Must be one of: staging, meta, public")
		return
	}

	if len(data.ApiKey.ValueString()) == 0 {
		apiKeyEnvVar := data.ApiKeyEnvVar.ValueString()
		if len(apiKeyEnvVar) == 0 {
			apiKeyEnvVar = "LIGHTSTEP_API_KEY"
		}
		apiKeyEnv, ok := os.LookupEnv(apiKeyEnvVar)
		if !ok {
			resp.Diagnostics.Append(diag.NewErrorDiagnostic(
				"No api key found",
				fmt.Sprintf("'api_key_env_var' is set to %v - but no api key found.", apiKeyEnvVar),
			))
			return
		}
		data.ApiKey = types.StringValue(apiKeyEnv)
	}
	// TODO remove this code once `environment` is fully deprecated
	if len(data.ApiUrl.ValueString()) == 0 {
		if apiUrlEnvVar, ok := os.LookupEnv("LIGHTSTEP_API_URL"); ok {
			data.ApiUrl = types.StringValue(apiUrlEnvVar)
		} else if apiUrlOtherEnvVar, ok := os.LookupEnv("LIGHTSTEP_API_BASE_URL"); ok {
			data.ApiUrl = types.StringValue(apiUrlOtherEnvVar)
		} else {
			// get the url from the `environment` provider attribute
			if data.Environment.ValueString() == "public" {
				data.ApiUrl = types.StringValue("https://api.lightstep.com")
			} else {
				data.ApiUrl = types.StringValue(fmt.Sprintf("https://api-%v.lightstep.com", data.Environment.ValueString()))
			}
		}
	}

	httpClient := client.NewClientWithUserAgent(
		data.ApiKey.ValueString(),
		data.Organization.ValueString(),
		data.ApiUrl.ValueString(),
		fmt.Sprintf("%s/%s (terraform %s)", "terraform-provider-lightstep", version.ProviderVersion, meta.SDKVersionString()),
	)
	resp.ResourceData = httpClient
	resp.DataSourceData = httpClient
}

func (p *lightstepProvider) Resources(ctx context.Context) []func() resource.Resource {
	return []func() resource.Resource{
		func() resource.Resource {
			return &resourceUnifiedDashboard{}
		},
		func() resource.Resource {
			return &resourceUnifiedDashboard{}
		},
	}
}

func (p *lightstepProvider) DataSources(ctx context.Context) []func() datasource.DataSource {
	return []func() datasource.DataSource{}
}
