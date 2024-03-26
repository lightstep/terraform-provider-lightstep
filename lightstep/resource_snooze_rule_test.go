package lightstep

import (
	"context"
	"fmt"
	"testing"

	"github.com/lightstep/terraform-provider-lightstep/client"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
)

func TestOneTimeSnoozeRule(t *testing.T) {
	var rule client.SnoozeRuleWithID

	ruleConfig := `
resource "lightstep_snooze_rule" "snooze1" {
  project_name = "` + testProject + `"
  title = "Snooze Test"
  
  scope {
    basic {
      scope_filter {
        alert_ids = ["alert1", "alert2"]
      }
      scope_filter {
        label_predicate {
	  	  operator = "eq"
	  	  label {
	  	    key = "a"
	  	    value = "b"
          }
        }
      }
    }
  }
  
  schedule {
    one_time {
      timezone = "America/Los_Angeles"
      start_date_time = "2021-03-20T00:00:00"
      end_date_time = "2021-03-24T14:30:00"
	}
  }
}
`
	// Change the title and a nested label value
	updatedRuleConfig := `
resource "lightstep_snooze_rule" "snooze1" {
  project_name = "` + testProject + `"
  title = "Snooze Test Updated"
  
  scope {
    basic {
      scope_filter {
        alert_ids = ["alert1", "alert2"]
      }
      scope_filter {
        label_predicate {
	  	  operator = "eq"
	  	  label {
	  	    key = "a"
	  	    value = "changed"
          }
        }
      }
    }
  }
  
  schedule {
    one_time {
      timezone = "America/Los_Angeles"
      start_date_time = "2021-03-20T00:00:00"
      end_date_time = "2021-03-24T14:30:00"
	}
  }
}
`

	resourceName := "lightstep_snooze_rule.snooze1"
	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccMetricConditionDestroy,
		Steps: []resource.TestStep{
			{
				Config: ruleConfig,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckSnoozeRuleExists(resourceName, &rule),
					resource.TestCheckResourceAttr(resourceName, "title", "Snooze Test"),
					resource.TestCheckResourceAttr(resourceName, "scope.0.basic.0.scope_filter.0.alert_ids.0", "alert1"),
					resource.TestCheckResourceAttr(resourceName, "scope.0.basic.0.scope_filter.0.alert_ids.1", "alert2"),
					resource.TestCheckResourceAttr(resourceName, "scope.0.basic.0.scope_filter.1.label_predicate.0.operator", "eq"),
					resource.TestCheckResourceAttr(resourceName, "scope.0.basic.0.scope_filter.1.label_predicate.0.label.0.key", "a"),
					resource.TestCheckResourceAttr(resourceName, "scope.0.basic.0.scope_filter.1.label_predicate.0.label.0.value", "b"),
					resource.TestCheckResourceAttr(resourceName, "schedule.0.one_time.0.timezone", "America/Los_Angeles"),
					resource.TestCheckResourceAttr(resourceName, "schedule.0.one_time.0.start_date_time", "2021-03-20T00:00:00"),
					resource.TestCheckResourceAttr(resourceName, "schedule.0.one_time.0.end_date_time", "2021-03-24T14:30:00"),
				),
			},
			{
				Config: updatedRuleConfig,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckSnoozeRuleExists(resourceName, &rule),
					resource.TestCheckResourceAttr(resourceName, "title", "Snooze Test Updated"),
					resource.TestCheckResourceAttr(resourceName, "scope.0.basic.0.scope_filter.1.label_predicate.0.label.0.value", "changed"),
				),
			},
		},
	})
}

func TestRecurringSnoozeRule(t *testing.T) {
	var rule client.SnoozeRuleWithID

	ruleConfig := `
resource "lightstep_snooze_rule" "snooze1" {
  project_name = "` + testProject + `"
  title = "Snooze Test"
  
  scope {
    basic {
      scope_filter {
        alert_ids = ["alert1"]
      }
    }
  }
  
  schedule {
    recurring {
      timezone = "America/Los_Angeles"
      start_date = "2021-03-20"
      end_date = "2021-03-24"
      schedule {
        name = "my schedule"
        start_time = "13:05:00"
        duration_millis = 1800000
        cadence {
          days_of_week = "1,3"
        }
      }
	}
  }
}
`
	resourceName := "lightstep_snooze_rule.snooze1"
	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccMetricConditionDestroy,
		Steps: []resource.TestStep{
			{
				Config: ruleConfig,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckSnoozeRuleExists(resourceName, &rule),
					resource.TestCheckResourceAttr(resourceName, "title", "Snooze Test"),
					resource.TestCheckResourceAttr(resourceName, "scope.0.basic.0.scope_filter.0.alert_ids.0", "alert1"),
					resource.TestCheckResourceAttr(resourceName, "schedule.0.recurring.0.timezone", "America/Los_Angeles"),
					resource.TestCheckResourceAttr(resourceName, "schedule.0.recurring.0.start_date", "2021-03-20"),
					resource.TestCheckResourceAttr(resourceName, "schedule.0.recurring.0.end_date", "2021-03-24"),
					resource.TestCheckResourceAttr(resourceName, "schedule.0.recurring.0.schedule.0.name", "my schedule"),
					resource.TestCheckResourceAttr(resourceName, "schedule.0.recurring.0.schedule.0.start_time", "13:05:00"),
					resource.TestCheckResourceAttr(resourceName, "schedule.0.recurring.0.schedule.0.duration_millis", "1800000"),
					resource.TestCheckResourceAttr(resourceName, "schedule.0.recurring.0.schedule.0.cadence.0.days_of_week", "1,3"),
				),
			},
		},
	})
}

func TestRecurringSnoozeRuleWithOptionalFieldsOmitted(t *testing.T) {
	var rule client.SnoozeRuleWithID

	ruleConfig := `
resource "lightstep_snooze_rule" "snooze1" {
  project_name = "` + testProject + `"
  title = "Snooze Test"
  
  scope {
    basic {
      scope_filter {
        alert_ids = ["alert1"]
      }
    }
  }
  
  schedule {
    recurring {
      timezone = "America/Los_Angeles"
      start_date = "2021-03-20"
      schedule {
        start_time = "13:05:00"
        duration_millis = 1800000
        cadence {
          days_of_week = "1,3"
        }
      }
	}
  }
}
`
	resourceName := "lightstep_snooze_rule.snooze1"
	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccMetricConditionDestroy,
		Steps: []resource.TestStep{
			{
				Config: ruleConfig,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckSnoozeRuleExists(resourceName, &rule),
					resource.TestCheckResourceAttr(resourceName, "title", "Snooze Test"),
					resource.TestCheckResourceAttr(resourceName, "scope.0.basic.0.scope_filter.0.alert_ids.0", "alert1"),
					resource.TestCheckResourceAttr(resourceName, "schedule.0.recurring.0.timezone", "America/Los_Angeles"),
					resource.TestCheckResourceAttr(resourceName, "schedule.0.recurring.0.start_date", "2021-03-20"),
					resource.TestCheckResourceAttr(resourceName, "schedule.0.recurring.0.schedule.0.start_time", "13:05:00"),
					resource.TestCheckResourceAttr(resourceName, "schedule.0.recurring.0.schedule.0.duration_millis", "1800000"),
					resource.TestCheckResourceAttr(resourceName, "schedule.0.recurring.0.schedule.0.cadence.0.days_of_week", "1,3"),
				),
			},
		},
	})
}

func testAccCheckSnoozeRuleExists(resourceName string, rule *client.SnoozeRuleWithID) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		tfSnoozeRule, ok := s.RootModule().Resources[resourceName]
		if !ok {
			return fmt.Errorf("not found: %s", resourceName)
		}

		if tfSnoozeRule.Primary.ID == "" {
			return fmt.Errorf("id is not set")
		}

		providerClient := testAccProvider.Meta().(*client.Client)
		r, err := providerClient.GetSnoozeRule(context.Background(), testProject, tfSnoozeRule.Primary.ID)
		if err != nil {
			return err
		}

		*rule = *r
		return nil
	}
}
