---
page_title: "lightstep_inferred_service_rule Resource - terraform-provider-lightstep"
subcategory: ""
description: |-

---

# lightstep_inferred_service_rule (Resource)

~> 🚧 This resource is under development and is not generally available yet. 🚧

Provides a [Lightstep Inferred Service Rule](https://docs.lightstep.com/docs/view-service-hierarchy-and-performance#add-inferred-services) that can detect and identify inferred services.


## Example Usage

```hcl
resource "lightstep_inferred_service_rule" "databases" {
  project_name = var.project
  name         = "database"
  description  = "Identifies select database management systems in the larger service topology"

  attribute_filters {
    key    = "span.kind"
    values = ["client"]
  }

  attribute_filters {
    key    = "db.type"
    values = ["sql", "redis", "memcached", "cassandra"]
  }

  group_by_keys = ["db.type", "db.instance"]
}
```

<!-- schema generated by tfplugindocs -->
## Schema

### Required

- `attribute_filters` (Block Set, Min: 1) Attribute filters that are checked against a leaf span's attributes to indicate the presence of the inferred service (see [below for nested schema](#nestedblock--attribute_filters))
- `name` (String) The name of the inferred service rule, which is included in the name of each inferred service created from this rule
- `project_name` (String) The name of the project to which the inferred service rule will apply

### Optional

- `description` (String) A description of the rule and what services it should infer
- `group_by_keys` (List of String) Attribute keys whose values will be included in the inferred service name

### Read-Only

- `id` (String) The ID of this resource.

<a id="nestedblock--attribute_filters"></a>
### Nested Schema for `attribute_filters`

Required:

- `key` (String) Key of a span attribute
- `values` (Set of String) Values for the attribute