package lightstep

import "github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"

func mergeSchemas(arr ...map[string]*schema.Schema) map[string]*schema.Schema {
	dst := make(map[string]*schema.Schema)
	for _, m := range arr {
		for k, v := range m {
			dst[k] = v
		}
	}
	return dst
}
