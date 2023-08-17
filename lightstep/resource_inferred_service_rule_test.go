package lightstep

import (
	"context"
	"fmt"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/lightstep/terraform-provider-lightstep/client"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

func TestAccInferredServiceRule(t *testing.T) {
	validInitialConfiguration := `
resource "lightstep_inferred_service_rule" "test_database_rule" {
	project_name = "` + testProject + `"
	name         = "database"
    description  =  "detects selected databases"

    attribute_filters {
		key    = "db.type"
		values = ["redis", "sql", "cassandra"]
    }

    attribute_filters {
		key    = "span.kind"
		values = ["client"]
    }

	group_by_keys = ["db.type", "db.instance"]
}

resource "lightstep_inferred_service_rule" "test_kafka_rule" {
	project_name = "` + testProject + `"
	name         = "kafka"

   attribute_filters {
		key    = "messaging.destination_kind"
		values = ["topic"]
   }
}
`

	validUpdatedConfiguration := `
resource "lightstep_inferred_service_rule" "test_database_rule" {
	project_name = "` + testProject + `"
	name         = "database"
    description  =  "detects selected databases"

    attribute_filters {
		key    = "db.type"
		values = ["memcached", "sql", "cassandra", "mongo"]
    }

	group_by_keys = ["db.type", "db.instance", "db.table.name"]
}	
`

	resource.Test(t, resource.TestCase{
		PreCheck: func() { testAccPreCheck(t) },
		// The deprecated Providers is required over ProviderFactories to initialize the provider's HTTP client for use
		// in `testAccInferredServiceRuleDestroy` and `testAccCheckInferredServiceRuleExists`
		Providers:    testAccProviders,
		CheckDestroy: testAccInferredServiceRuleDestroy,
		Steps: []resource.TestStep{
			{
				Config: validInitialConfiguration,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckInferredServiceRuleExists("lightstep_inferred_service_rule.test_database_rule"),
					resource.TestCheckResourceAttr("lightstep_inferred_service_rule.test_database_rule", "project_name", testProject),
					resource.TestCheckResourceAttr("lightstep_inferred_service_rule.test_database_rule", "name", "database"),
					resource.TestCheckResourceAttr("lightstep_inferred_service_rule.test_database_rule", "description", "detects selected databases"),
					resource.TestCheckResourceAttr("lightstep_inferred_service_rule.test_database_rule", "attribute_filters.#", "2"),
					resource.TestCheckTypeSetElemNestedAttrs("lightstep_inferred_service_rule.test_database_rule", "attribute_filters.*", map[string]string{"key": "db.type", "values.#": "3"}),
					resource.TestCheckTypeSetElemAttr("lightstep_inferred_service_rule.test_database_rule", "attribute_filters.*.values.*", "cassandra"),
					resource.TestCheckTypeSetElemAttr("lightstep_inferred_service_rule.test_database_rule", "attribute_filters.*.values.*", "redis"),
					resource.TestCheckTypeSetElemAttr("lightstep_inferred_service_rule.test_database_rule", "attribute_filters.*.values.*", "sql"),
					resource.TestCheckTypeSetElemNestedAttrs("lightstep_inferred_service_rule.test_database_rule", "attribute_filters.*", map[string]string{"key": "span.kind", "values.#": "1"}),
					resource.TestCheckTypeSetElemAttr("lightstep_inferred_service_rule.test_database_rule", "attribute_filters.*.values.*", "client"),
					resource.TestCheckResourceAttr("lightstep_inferred_service_rule.test_database_rule", "group_by_keys.#", "2"),
					resource.TestCheckResourceAttr("lightstep_inferred_service_rule.test_database_rule", "group_by_keys.0", "db.type"),
					resource.TestCheckResourceAttr("lightstep_inferred_service_rule.test_database_rule", "group_by_keys.1", "db.instance"),

					testAccCheckInferredServiceRuleExists("lightstep_inferred_service_rule.test_kafka_rule"),
					resource.TestCheckResourceAttr("lightstep_inferred_service_rule.test_kafka_rule", "project_name", testProject),
					resource.TestCheckResourceAttr("lightstep_inferred_service_rule.test_kafka_rule", "name", "kafka"),
					resource.TestCheckResourceAttr("lightstep_inferred_service_rule.test_kafka_rule", "description", ""),
					resource.TestCheckResourceAttr("lightstep_inferred_service_rule.test_kafka_rule", "attribute_filters.#", "1"),
					resource.TestCheckTypeSetElemNestedAttrs("lightstep_inferred_service_rule.test_kafka_rule", "attribute_filters.*", map[string]string{"key": "messaging.destination_kind"}),
					resource.TestCheckTypeSetElemAttr("lightstep_inferred_service_rule.test_kafka_rule", "attribute_filters.*.values.*", "topic"),
					resource.TestCheckResourceAttr("lightstep_inferred_service_rule.test_kafka_rule", "group_by_keys.#", "0"),
				),
			},
			{
				Config: validUpdatedConfiguration,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("lightstep_inferred_service_rule.test_database_rule", "project_name", testProject),
					resource.TestCheckResourceAttr("lightstep_inferred_service_rule.test_database_rule", "name", "database"),
					resource.TestCheckResourceAttr("lightstep_inferred_service_rule.test_database_rule", "description", "detects selected databases"),
					resource.TestCheckResourceAttr("lightstep_inferred_service_rule.test_database_rule", "attribute_filters.#", "1"),
					resource.TestCheckTypeSetElemNestedAttrs("lightstep_inferred_service_rule.test_database_rule", "attribute_filters.*", map[string]string{"key": "db.type", "values.#": "4"}),
					resource.TestCheckTypeSetElemAttr("lightstep_inferred_service_rule.test_database_rule", "attribute_filters.*.values.*", "memcached"),
					resource.TestCheckTypeSetElemAttr("lightstep_inferred_service_rule.test_database_rule", "attribute_filters.*.values.*", "cassandra"),
					resource.TestCheckTypeSetElemAttr("lightstep_inferred_service_rule.test_database_rule", "attribute_filters.*.values.*", "sql"),
					resource.TestCheckTypeSetElemAttr("lightstep_inferred_service_rule.test_database_rule", "attribute_filters.*.values.*", "mongo"),
					resource.TestCheckResourceAttr("lightstep_inferred_service_rule.test_database_rule", "group_by_keys.#", "3"),
					resource.TestCheckResourceAttr("lightstep_inferred_service_rule.test_database_rule", "group_by_keys.0", "db.type"),
					resource.TestCheckResourceAttr("lightstep_inferred_service_rule.test_database_rule", "group_by_keys.1", "db.instance"),
					resource.TestCheckResourceAttr("lightstep_inferred_service_rule.test_database_rule", "group_by_keys.2", "db.table.name"),
				),
			},
		},
	})
}

func TestAccInferredServiceRuleImport(t *testing.T) {
	importConfig := `
resource "lightstep_inferred_service_rule" "test_import_rule_db" {
    project_name = "` + testProject + `"
    name         = "database"
   description  =  "detects selected databases"

   attribute_filters {
        key    = "db.type"
        values = ["redis", "sql", "cassandra"]
   }

   attribute_filters {
        key    = "span.kind"
        values = ["client"]
   }

    group_by_keys = ["db.type", "db.instance"]
}

resource "lightstep_inferred_service_rule" "test_import_rule_kafka" {
    project_name = "` + testProject + `"
    name         = "kafka"

  attribute_filters {
        key    = "messaging.destination_kind"
        values = ["topic"]
  }
}
`

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: testAccProviderFactories,
		CheckDestroy:      testAccInferredServiceRuleDestroy,
		Steps: []resource.TestStep{
			{
				Config: importConfig,
			},
			{
				ResourceName:        "lightstep_inferred_service_rule.test_import_rule_db",
				ImportState:         true,
				ImportStateVerify:   true,
				ImportStateIdPrefix: fmt.Sprintf("%s.", testProject),
			},
			{
				ResourceName:        "lightstep_inferred_service_rule.test_import_rule_kafka",
				ImportState:         true,
				ImportStateVerify:   true,
				ImportStateIdPrefix: fmt.Sprintf("%s.", testProject),
			},
		},
	})
}

func testAccCheckInferredServiceRuleExists(resourceName string) resource.TestCheckFunc {
	return func(tfState *terraform.State) error {
		tfResource, ok := tfState.RootModule().Resources[resourceName]
		if !ok {
			return fmt.Errorf("not found: %s", resourceName)
		}

		if tfResource.Primary.ID == "" {
			return fmt.Errorf("id is not set")
		}

		apiClient := testAccProvider.Meta().(*client.Client)
		_, err := apiClient.GetInferredServiceRule(context.Background(), testProject, tfResource.Primary.ID)
		if err != nil {
			return err
		}

		return nil
	}
}

func testAccInferredServiceRuleDestroy(tfState *terraform.State) error {
	apiClient := testAccProvider.Meta().(*client.Client)
	for _, tfResource := range tfState.RootModule().Resources {
		if tfResource.Type != "lightstep_inferred_service_rule" {
			continue
		}

		apiResponse, err := apiClient.GetInferredServiceRule(context.Background(), testProject, tfResource.Primary.ID)
		if err == nil {
			if apiResponse.ID == tfResource.Primary.ID {
				return fmt.Errorf("inferred service rule with ID (%v) still exists", tfResource.Primary.ID)
			}
		}

	}
	return nil
}
