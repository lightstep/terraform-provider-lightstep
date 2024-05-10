package lightstep

import (
	"context"
	"fmt"
	"github.com/hashicorp/terraform-plugin-framework-validators/helpers/validatordiag"
	"github.com/hashicorp/terraform-plugin-framework-validators/int64validator"
	"github.com/hashicorp/terraform-plugin-framework-validators/setvalidator"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	schema2 "github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/lightstep/terraform-provider-lightstep/client"
)

func getUnifiedQuerySchemaMap() map[string]schema2.Attribute {
	sma := map[string]schema2.Attribute{
		"hidden": schema2.BoolAttribute{
			Required: true,
		},
		"display": schema2.StringAttribute{
			Optional: true,
			Validators: []validator.String{
				stringvalidator.OneOf(
					"line",
					"area",
					"bar",
					"big_number",
					"heatmap",
					"dependency_map",
					"big_number_v2",
					"scatter_plot",
					"ordered_list",
					"pie",
					"table",
					"traces_list",
				),
			},
		},
		// See https://github.com/hashicorp/terraform-plugin-sdk/issues/155
		// Using a TypeSet of size 1 as a way to allow nested properties
		"display_type_options": schema2.SetNestedAttribute{
			Description: "Applicable options vary depending on the display type. Please see the Lightstep documentation for a full description.",
			Optional:    true,
			Validators: []validator.Set{
				setvalidator.SizeAtMost(1),
			},
			NestedObject: schema2.NestedAttributeObject{
				Attributes: map[string]schema2.Attribute{
					"display_type": schema2.StringAttribute{
						Optional: true,
					},
					"sort_by": schema2.StringAttribute{
						Optional: true,
					},
					"sort_direction": schema2.StringAttribute{
						Optional: true,
					},
					"y_axis_scale": schema2.StringAttribute{
						Optional: true,
						Validators: []validator.String{
							stringvalidator.OneOf(
								"linear",
								"log",
								"symlog",
							),
						},
					},
					"y_axis_log_base": schema2.Int64Attribute{
						Optional: true,
						Validators: []validator.Int64{
							int64validator.OneOf(2, 10),
						},
					},
					"y_axis_min": schema2.Float64Attribute{
						Optional: true,
					},
					"y_axis_max": schema2.Float64Attribute{
						Optional: true,
					},
					"is_donut": schema2.BoolAttribute{
						Optional: true,
					},
					"comparison_window_ms": schema2.Int64Attribute{
						Optional: true,
					},
				},
			},
		},
		"query_name": schema2.StringAttribute{
			Required: true,
			Validators: []validator.String{
				stringvalidator.LengthBetween(1, 128),
			},
		},
		"query_string": schema2.StringAttribute{
			Required: true,
		},
		"hidden_queries": schema2.MapAttribute{
			ElementType: types.StringType,
			Optional:    true,
			Description: "An optional map of sub-query names in the query_string to a boolean string to hide/show that query. " +
				"If specified, the map must have an entry for all named sub-queries in the query_string. A value " +
				"of \"true\" indicates the query should be hidden. " +
				"Example: `hidden_queries = {  \"a\" = \"true\",  \"b\" = \"false\" }`.",
			Validators: []validator.Map{
				newHiddenQueriesValidator(),
			},
		},
	}
	return sma
}

var _ validator.Map = hiddenQueriesValidator{}

// hiddenQueriesValidator validates that the top-level query's name is not specified in hidden_queries
type hiddenQueriesValidator struct{}

// Description describes the validation in plain text formatting.
func (v hiddenQueriesValidator) Description(_ context.Context) string {
	return "hidden_queries must not contain top-level query name"
}

// MarkdownDescription describes the validation in Markdown formatting.
func (v hiddenQueriesValidator) MarkdownDescription(ctx context.Context) string {
	return v.Description(ctx)
}

// Validate performs the validation.
func (v hiddenQueriesValidator) ValidateMap(ctx context.Context, req validator.MapRequest, resp *validator.MapResponse) {
	var topLevelQueryName string
	d := req.Config.GetAttribute(ctx, req.Path.ParentPath().AtName("query_name"), &topLevelQueryName)
	if d.HasError() {
		resp.Diagnostics.Append(d...)
		return
	}
	for queryName := range req.ConfigValue.Elements() {
		if queryName == topLevelQueryName {
			resp.Diagnostics.Append(validatordiag.InvalidAttributeValueDiagnostic(
				req.Path,
				v.Description(ctx),
				fmt.Sprintf("%v", req.ConfigValue.Elements()),
			))
		}
	}
}

func newHiddenQueriesValidator() validator.Map {
	return hiddenQueriesValidator{}
}

func getQueriesFromUnifiedDashboardResourceData(
	queriesIn []client.MetricQueryWithAttributes,
	dashboardID string,
	chartID string,
) ([]interface{}, error) {

	var queries []interface{}
	for _, q := range queriesIn {
		// Check if a query is using the legacy representation.  In that case,
		// it can't be represented in the "query_string" field and the call should
		// fail. In the future, we should attempt to automatically convert legacy -> query
		// string format. At the moment, there's no public API to implement this, so
		// at least provide a clarifying error message.
		if q.Type != "tql" {
			return nil, fmt.Errorf(
				"cannot convert query from chart %v in dashboard %v\n\n"+
					"Query is of type '%v' but must be of type 'tql' for use with the resource\n"+
					"type lightstep_dashboard.\n"+
					"\n"+
					"Try using the lightstep_metrics_dashboard resource type for this dashboard\n"+
					"or convert all queries in the dashboard to query string format. ",
				chartID,
				dashboardID,
				q.Type,
			)
		}

		qs := map[string]interface{}{
			"hidden":                 q.Hidden,
			"display":                q.Display,
			"display_type_options":   convertNestedMapToSchemaSet(q.DisplayTypeOptions),
			"query_name":             q.Name,
			"query_string":           q.QueryString,
			"dependency_map_options": getDependencyMapOptions(q.DependencyMapOptions),
		}
		setHiddenQueriesFromResourceData(qs, q)

		queries = append(queries, qs)
	}
	return queries, nil
}

func getDependencyMapOptions(options *client.DependencyMapOptions) []interface{} {
	if options == nil {
		return nil
	}

	return []interface{}{
		map[string]interface{}{
			"map_type": options.MapType,
			"scope":    options.Scope,
		},
	}
}
