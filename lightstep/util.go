package lightstep

import (
	"fmt"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/lightstep/terraform-provider-lightstep/client"
)

func mergeSchemas(arr ...map[string]*schema.Schema) map[string]*schema.Schema {
	dst := make(map[string]*schema.Schema)
	for _, m := range arr {
		for k, v := range m {
			dst[k] = v
		}
	}
	return dst
}

// takes a nested map (such as display_type_options or panel_options) and converts to a schema set
func convertNestedMapToSchemaSet(opts map[string]interface{}) *schema.Set {
	// nested maps contain a set that always has at most one element, so
	// the hash function is trivial
	f := func(i interface{}) int {
		return 1
	}
	return schema.NewSet(f, []interface{}{opts})
}

// Note due to Terraform's issues with TypeMap having TypeBool elements, we
// need to use boolean strings
func setHiddenQueriesFromResourceData(qs map[string]interface{}, query client.MetricQueryWithAttributes) {
	if len(query.HiddenQueries) == 0 {
		// nothing to do
		return
	}
	hq := make(map[string]interface{}, len(query.HiddenQueries))
	for k, v := range query.HiddenQueries {
		hq[k] = fmt.Sprintf("%t", v)
	}
	qs["hidden_queries"] = hq
}
