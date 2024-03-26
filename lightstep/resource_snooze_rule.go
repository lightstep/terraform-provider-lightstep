package lightstep

import (
	"context"
	"fmt"
	"net/http"
	"strings"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"

	"github.com/lightstep/terraform-provider-lightstep/client"
)

// resourceSnoozeRule creates a resource for either:
//
// (1) The legacy lightstep_metric_condition
// (2) The unified lightstep_alert
//
// The resources are largely the same with the primary differences being the
// query format and composite alert support.
func resourceSnoozeRule() *schema.Resource {
	p := resourceSnoozeRuleImp{}

	resource := &schema.Resource{
		CreateContext: p.resourceSnoozeRuleCreate,
		ReadContext:   p.resourceSnoozeRuleRead,
		UpdateContext: p.resourceSnoozeRuleUpdate,
		DeleteContext: p.resourceSnoozeRuleDelete,
		Importer: &schema.ResourceImporter{
			StateContext: p.resourceSnoozeRuleImport,
		},
		Schema: map[string]*schema.Schema{
			"project_name": {
				Type:        schema.TypeString,
				Required:    true,
				Description: "The name of the [project](https://docs.lightstep.com/docs/glossary#project) in which to create this alert.",
			},
			"title": {
				Type:        schema.TypeString,
				Required:    true,
				Description: "The title of the snooze rule.",
			},
			"scope": {
				Type:        schema.TypeSet,
				Required:    true,
				MinItems:    1,
				MaxItems:    1,
				Description: "Defines which alerts the rule applies to",
				Elem: &schema.Resource{
					Schema: getScopeSchemaMap(),
				},
			},
			"schedule": {
				Type:        schema.TypeSet,
				Required:    true,
				Description: "Defines when the silencing rule is effective",
				MinItems:    1,
				MaxItems:    1,
				Elem: &schema.Resource{
					Schema: getScheduleSchemaMap(),
				},
			},
		},
	}
	return resource
}

func getScheduleSchemaMap() map[string]*schema.Schema {
	return map[string]*schema.Schema{
		"one_time": {
			Type:        schema.TypeSet,
			Optional:    true,
			MinItems:    1,
			MaxItems:    1,
			Description: "Effective during the entire specified window",
			Elem: &schema.Resource{
				Schema: getOneTimeScheduleSchemaMap(),
			},
		},
		"recurring": {
			Type:        schema.TypeSet,
			Optional:    true,
			MinItems:    1,
			MaxItems:    1,
			Description: "Effective beginning at the start date and follows the schedules defined. When schedules overlap, the rule is effective",
			Elem: &schema.Resource{
				Schema: getRecurringScheduleSchemaMap(),
			},
		},
	}
}

func getRecurringScheduleSchemaMap() map[string]*schema.Schema {
	return map[string]*schema.Schema{
		"timezone": {
			Type:        schema.TypeString,
			Required:    true,
			Description: "IANA format timezone. Examples: 'UTC', 'US/Pacific', 'Europe/Paris'",
		},
		"start_date": {
			Type:        schema.TypeString,
			Required:    true,
			Description: "ISO 8601 date format. Example: 2021-01-01",
		},
		"end_date": {
			Type:        schema.TypeString,
			Optional:    true,
			Description: "ISO 8601 date format. Example: 2021-01-01",
		},
		"schedule": {
			Type:        schema.TypeSet,
			Required:    true,
			MinItems:    1,
			MaxItems:    1,
			Description: "",
			Elem: &schema.Resource{
				Schema: getReoccurrenceSchemaMap(),
			},
		},
	}
}

func getReoccurrenceSchemaMap() map[string]*schema.Schema {
	return map[string]*schema.Schema{
		"name": {
			Type:        schema.TypeString,
			Optional:    true,
			Description: "Human-readable name for this reoccurrence",
		},
		"start_time": {
			Type:        schema.TypeString,
			Required:    true,
			Description: "ISO 8601 time format defining when the silencing period begins on each relevant day defined by the cadence. Must NOT include UTC time offset (the time zone is specified in the 'recurring' block instead. Example '16:07:29'",
		},
		"duration_millis": {
			Type:        schema.TypeInt,
			Required:    true,
			Description: "How long each occurrence lasts specified in milliseconds. Must be a multiple of 1 minute (no fractional minutes)",
		},
		"cadence": {
			Type:        schema.TypeSet,
			Required:    true,
			MinItems:    1,
			MaxItems:    1,
			Description: "Defines which days should have an instance of this reoccurrence",
			Elem: &schema.Resource{
				Schema: getCadenceSchemaMap(),
			},
		},
	}
}

func getCadenceSchemaMap() map[string]*schema.Schema {
	return map[string]*schema.Schema{
		"days_of_week": {
			Type:        schema.TypeString,
			Required:    true,
			Description: "Comma-separated List of number or ranges (crontab-style). The empty string is defined as no days. Leaving this field undefined or null is defined as all days.a The string '*' is also defined as all days. Format: 0, 7 = sun, 1 = mon, ..., 6 = stat. Examples: '1-5' or '6-7' or '2,4', '*', ''",
		},
	}
}

func getOneTimeScheduleSchemaMap() map[string]*schema.Schema {
	return map[string]*schema.Schema{
		"timezone": {
			Type:        schema.TypeString,
			Required:    true,
			Description: "IANA format timezone. Examples: 'UTC', 'US/Pacific', 'Europe/Paris'",
		},
		"start_date_time": {
			Type:        schema.TypeString,
			Required:    true,
			Description: "ISO 8601 relative date/time format. Example: '2021-04-04T14:30:00'",
		},
		"end_date_time": {
			Type:        schema.TypeString,
			Optional:    true,
			Description: "ISO 8601 relative date/time format. Example: '2021-04-04T14:30:00'",
		},
	}
}

func getScopeSchemaMap() map[string]*schema.Schema {
	return map[string]*schema.Schema{
		"basic": {
			Type:        schema.TypeSet,
			Required:    true,
			MinItems:    1,
			MaxItems:    1,
			Description: "Defines which alerts the rule applies to",
			Elem: &schema.Resource{
				Schema: getBasicTargetingSchemaMap(),
			},
		},
	}
}

func getBasicTargetingSchemaMap() map[string]*schema.Schema {
	return map[string]*schema.Schema{
		"scope_filter": {
			Type:        schema.TypeSet,
			Required:    true,
			Description: "Defines which alerts the rule applies to",
			MinItems:    1,
			Elem: &schema.Resource{
				Schema: getScopeFilterSchemaMap(),
			},
		},
	}
}

func getScopeFilterSchemaMap() map[string]*schema.Schema {
	return map[string]*schema.Schema{
		"alert_ids": {
			Type:     schema.TypeSet,
			Optional: true,
			Elem: &schema.Schema{
				Type: schema.TypeString,
			},
		},
		"label_predicate": {
			Type:        schema.TypeSet,
			Optional:    true,
			Description: "Optional configuration to receive alert notifications.",
			MinItems:    1,
			MaxItems:    1,
			Elem: &schema.Resource{
				Schema: getPredicateSchemaMap(),
			},
		},
	}
}

type resourceSnoozeRuleImp struct {
}

func (p *resourceSnoozeRuleImp) resourceSnoozeRuleCreate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	c := m.(*client.Client)
	rule, err := getSnoozeRuleFromResource(d)
	if err != nil {
		return diag.FromErr(fmt.Errorf("failed to get snooze rule from resource : %v", err))
	}

	created, err := c.CreateSnoozeRule(ctx, d.Get("project_name").(string), rule)
	if err != nil {
		return diag.FromErr(fmt.Errorf("failed to create snooze rule: %v", err))
	}

	d.SetId(created.ID)

	return p.resourceSnoozeRuleRead(ctx, d, m)
}

func (p *resourceSnoozeRuleImp) resourceSnoozeRuleRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	var diags diag.Diagnostics

	c := m.(*client.Client)
	projectName := d.Get("project_name").(string)
	rule, err := c.GetSnoozeRule(ctx, projectName, d.Id())
	if err != nil {
		apiErr, ok := err.(client.APIResponseCarrier)
		if !ok {
			return diag.FromErr(fmt.Errorf("failed to get snooze rule: %v", err))
		}

		if apiErr.GetStatusCode() == http.StatusNotFound {
			d.SetId("")
			return diags
		}

		return diag.FromErr(fmt.Errorf("failed to get snooze rule: %v", apiErr))
	}

	err = setResourceDataFromSnoozeRule(projectName, *rule, d)
	if err != nil {
		return diag.FromErr(fmt.Errorf("failed to set snooze rule from API response to terraform state: %v", err))
	}

	return diags
}

func (p *resourceSnoozeRuleImp) resourceSnoozeRuleUpdate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	c := m.(*client.Client)
	rule, err := getSnoozeRuleFromResource(d)
	if err != nil {
		return diag.FromErr(fmt.Errorf("failed to get snooze rule from resource : %v", err))
	}

	if _, err := c.UpdateSnoozeRule(ctx, d.Get("project_name").(string), d.Id(), rule); err != nil {
		return diag.FromErr(fmt.Errorf("failed to update snooze rule: %v", err))
	}

	return p.resourceSnoozeRuleRead(ctx, d, m)
}

func (p *resourceSnoozeRuleImp) resourceSnoozeRuleDelete(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	var diags diag.Diagnostics

	c := m.(*client.Client)
	if err := c.DeleteSnoozeRule(ctx, d.Get("project_name").(string), d.Id()); err != nil {
		return diag.FromErr(fmt.Errorf("failed to delete snooze rule: %v", err))
	}

	// d.SetId("") is automatically called assuming delete returns no errors, but
	// it is added here for explicitness.
	d.SetId("")
	return diags
}

func (p *resourceSnoozeRuleImp) resourceSnoozeRuleImport(ctx context.Context, d *schema.ResourceData, m interface{}) ([]*schema.ResourceData, error) {
	c := m.(*client.Client)

	ids := strings.Split(d.Id(), ".")
	if len(ids) != 2 {
		resourceName := "lightstep_snooze_rule"
		return []*schema.ResourceData{}, fmt.Errorf("error importing %v. Expecting an  ID formed as '<lightstep_project>.<%v_ID>'", resourceName, resourceName)
	}

	project, id := ids[0], ids[1]
	rule, err := c.GetSnoozeRule(ctx, project, id)
	if err != nil {
		return []*schema.ResourceData{}, fmt.Errorf("failed to get metric condition. err: %v", err)
	}

	d.SetId(id)
	if err := setResourceDataFromSnoozeRule(project, *rule, d); err != nil {
		return nil, fmt.Errorf("failed to set snooze rule from API response to terraform state: %v", err)
	}

	return []*schema.ResourceData{d}, nil
}

func setResourceDataFromSnoozeRule(project string, rule client.SnoozeRuleWithID, d *schema.ResourceData) error {
	d.SetId(rule.ID)
	err := d.Set("title", rule.Title)
	if err != nil {
		return fmt.Errorf("failed to set title: %v", err)
	}

	scopeMap, err := scopeToMap(rule.Scope)
	if err != nil {
		return fmt.Errorf("failed to convert scope: %v", err)
	}
	err = d.Set("scope", schema.NewSet(
		schema.HashResource(&schema.Resource{
			Schema: getScopeSchemaMap(),
		}), []any{scopeMap},
	))
	if err != nil {
		return fmt.Errorf("failed to set scope map: %v", err)
	}

	scheduleMap, err := scheduleToMap(rule.Schedule)
	if err != nil {
		return fmt.Errorf("failed to convert schedule: %v", err)
	}
	err = d.Set("schedule", schema.NewSet(
		schema.HashResource(&schema.Resource{
			Schema: getScheduleSchemaMap(),
		}), []any{scheduleMap},
	))
	if err != nil {
		return fmt.Errorf("failed to set schedule map: %v", err)
	}
	return nil
}

func scheduleToMap(schedule client.Schedule) (map[string]any, error) {
	scheduleMap := make(map[string]any)
	if schedule.OneTime != nil {
		oneTimeMap, err := oneTimeScheduleToMap(*schedule.OneTime)
		if err != nil {
			return nil, fmt.Errorf("failed to convert one time schedule: %v", err)
		}

		scheduleMap["one_time"] = schema.NewSet(
			schema.HashResource(&schema.Resource{
				Schema: getOneTimeScheduleSchemaMap(),
			}), []any{oneTimeMap},
		)
	}
	if schedule.Recurring != nil {
		oneTimeMap, err := recurringScheduleToMap(*schedule.Recurring)
		if err != nil {
			return nil, fmt.Errorf("failed to convert one time schedule: %v", err)
		}

		scheduleMap["recurring"] = schema.NewSet(
			schema.HashResource(&schema.Resource{
				Schema: getRecurringScheduleSchemaMap(),
			}), []any{oneTimeMap},
		)
	}
	return scheduleMap, nil
}

func oneTimeScheduleToMap(oneTimeSchedule client.OneTimeSchedule) (map[string]any, error) {
	oneTimeScheduleMap := make(map[string]any)
	oneTimeScheduleMap["timezone"] = oneTimeSchedule.Timezone
	oneTimeScheduleMap["start_date_time"] = oneTimeSchedule.StartDateTime
	oneTimeScheduleMap["end_date_time"] = oneTimeSchedule.EndDateTime
	return oneTimeScheduleMap, nil
}

func recurringScheduleToMap(recurringSchedule client.RecurringSchedule) (map[string]any, error) {
	recurringScheduleMap := make(map[string]any)

	recurringScheduleMap["timezone"] = recurringSchedule.Timezone
	recurringScheduleMap["start_date"] = recurringSchedule.StartDate
	recurringScheduleMap["end_date"] = recurringSchedule.EndDate

	var reocurrences []any
	for _, s := range recurringSchedule.Schedules {
		reocurrenceMap, err := reoccurrenceToMap(s)
		if err != nil {
			return nil, fmt.Errorf("failed to convert recurring schedule: %v", err)
		}
		reocurrences = append(reocurrences, reocurrenceMap)
	}
	recurringScheduleMap["schedule"] = schema.NewSet(
		schema.HashResource(&schema.Resource{
			Schema: getReoccurrenceSchemaMap(),
		}), reocurrences,
	)
	return recurringScheduleMap, nil
}

func reoccurrenceToMap(reoccurrence client.Reoccurrence) (map[string]any, error) {
	reoccurrenceMap := make(map[string]any)
	reoccurrenceMap["name"] = reoccurrence.Name
	reoccurrenceMap["start_time"] = reoccurrence.StartTime
	reoccurrenceMap["duration_millis"] = int(reoccurrence.DurationMillis)

	cadenceMap, err := cadenceToMap(reoccurrence.Cadence)
	if err != nil {
		return nil, fmt.Errorf("failed to convert cadence: %v", err)
	}
	reoccurrenceMap["cadence"] = schema.NewSet(
		schema.HashResource(&schema.Resource{
			Schema: getCadenceSchemaMap(),
		}), []any{cadenceMap},
	)
	return reoccurrenceMap, nil
}

func cadenceToMap(cadence client.Cadence) (map[string]any, error) {
	cadenceMap := make(map[string]any)
	cadenceMap["days_of_week"] = cadence.DaysOfWeek
	return cadenceMap, nil
}

func scopeToMap(scope client.Scope) (map[string]any, error) {
	scopeMap := make(map[string]any)
	if scope.Basic != nil {
		basicTargeting, err := basicTargetingToMap(*scope.Basic)
		if err != nil {
			return nil, fmt.Errorf("failed to convert basic targeting: %v", err)
		}

		scopeMap["basic"] = schema.NewSet(
			schema.HashResource(&schema.Resource{
				Schema: getBasicTargetingSchemaMap(),
			}), []any{basicTargeting},
		)
	}
	return scopeMap, nil
}

func basicTargetingToMap(basicTargeting client.BasicTargeting) (map[string]any, error) {
	var scopeFilterMaps []any
	for _, scopeFilter := range basicTargeting.ScopeFilters {
		scopeFilterMap, err := scopeFiltersToMap(scopeFilter)
		if err != nil {
			return nil, fmt.Errorf("failed to convert basic targeting: %v", err)
		}
		scopeFilterMaps = append(scopeFilterMaps, scopeFilterMap)
	}

	basicTargetingMap := make(map[string]any)
	basicTargetingMap["scope_filter"] = schema.NewSet(
		schema.HashResource(&schema.Resource{
			Schema: getScopeFilterSchemaMap(),
		}), scopeFilterMaps,
	)
	return basicTargetingMap, nil
}

func scopeFiltersToMap(filter client.ScopeFilter) (map[string]any, error) {
	basicTargetingMap := make(map[string]any)

	predicateMap, err := predicateToMap(filter.LabelPredicate)
	if err != nil {
		return nil, fmt.Errorf("failed to convert predicate: %v", err)
	}

	aidSet := schema.NewSet(schema.HashString, nil)
	for _, aid := range filter.AlertIDs {
		aidSet.Add(aid)
	}
	basicTargetingMap["alert_ids"] = aidSet
	basicTargetingMap["label_predicate"] = schema.NewSet(
		schema.HashResource(&schema.Resource{
			Schema: getPredicateSchemaMap(),
		}), []any{predicateMap},
	)
	return basicTargetingMap, nil
}

func predicateToMap(predicate *client.Predicate) (map[string]any, error) {
	if predicate == nil {
		return nil, nil
	}

	predicateMap := make(map[string]any)
	predicateMap["operator"] = predicate.Operator

	var labelMaps []any
	for _, label := range predicate.Labels {
		labelMap, err := labelToMap(label)
		if err != nil {
			return nil, fmt.Errorf("failed to convert label: %v", err)
		}
		labelMaps = append(labelMaps, labelMap)
	}
	predicateMap["label"] = schema.NewSet(
		schema.HashResource(&schema.Resource{
			Schema: getLabelsSchemaMap(),
		}), labelMaps,
	)
	return predicateMap, nil
}

func labelToMap(label client.ResourceLabel) (map[string]any, error) {
	labelMap := make(map[string]any)
	labelMap["key"] = label.Key
	labelMap["value"] = label.Value
	return labelMap, nil
}

func getSnoozeRuleFromResource(d *schema.ResourceData) (client.SnoozeRule, error) {
	var snoozeRule client.SnoozeRule
	snoozeRule.Title = d.Get("title").(string)
	if schedule, ok := getFirst(d.Get("schedule")); ok {
		if oneTime, ok := getFirst(schedule["one_time"]); ok {
			snoozeRule.Schedule.OneTime = &client.OneTimeSchedule{
				Timezone:      oneTime["timezone"].(string),
				StartDateTime: oneTime["start_date_time"].(string),
				EndDateTime:   oneTime["end_date_time"].(string),
			}
		}
		if recurring, ok := getFirst(schedule["recurring"]); ok {
			var rs []client.Reoccurrence
			if reoccurrences, ok := getAll[map[string]any](recurring["schedule"]); ok {
				for _, reoccurrence := range reoccurrences {
					r := client.Reoccurrence{
						Name:           reoccurrence["name"].(string),
						StartTime:      reoccurrence["start_time"].(string),
						DurationMillis: int64(reoccurrence["duration_millis"].(int)),
					}
					if cadence, ok := getFirst(reoccurrence["cadence"]); ok {
						r.Cadence = client.Cadence{
							DaysOfWeek: cadence["days_of_week"].(string),
						}
					}
					rs = append(rs, r)
				}
			}
			snoozeRule.Schedule.Recurring = &client.RecurringSchedule{
				Timezone:  recurring["timezone"].(string),
				StartDate: recurring["start_date"].(string),
				EndDate:   recurring["end_date"].(string),
				Schedules: rs,
			}
		}
	}
	if scope, ok := getFirst(d.Get("scope")); ok {
		if basic, ok := getFirst(scope["basic"]); ok {
			if scopeFilters, ok := getAll[map[string]any](basic["scope_filter"]); ok {
				var sfs []client.ScopeFilter
				for _, scopeFilter := range scopeFilters {
					var sf client.ScopeFilter
					if alertIds, ok := getAll[string](scopeFilter["alert_ids"]); ok {
						sf.AlertIDs = alertIds
					}
					if labelPredicate, ok := getFirst(scopeFilter["label_predicate"]); ok {
						sf.LabelPredicate = &client.Predicate{
							Operator: labelPredicate["operator"].(string),
						}
						if labels, ok := getAll[map[string]any](labelPredicate["label"]); ok {
							for _, label := range labels {
								var l client.ResourceLabel
								l.Key = label["key"].(string)
								l.Value = label["value"].(string)
								sf.LabelPredicate.Labels = append(sf.LabelPredicate.Labels, l)
							}
						}
					}
					sfs = append(sfs, sf)
				}
				snoozeRule.Scope.Basic = &client.BasicTargeting{
					ScopeFilters: sfs,
				}
			}
		}
	}
	return snoozeRule, nil
}

func getFirst(x any) (map[string]any, bool) {
	if s, ok := x.(*schema.Set); ok {
		l := s.List()
		if len(l) > 0 {
			v, ok := l[0].(map[string]any)
			return v, ok
		}
	}
	return nil, false
}

func getAll[T any](x any) ([]T, bool) {
	var all []T
	if s, ok := x.(*schema.Set); ok {
		l := s.List()
		for _, i := range l {
			if v, ok := i.(T); ok {
				all = append(all, v)
			} else {
				return nil, false
			}
		}
	} else {
		return nil, false
	}
	return all, true
}
