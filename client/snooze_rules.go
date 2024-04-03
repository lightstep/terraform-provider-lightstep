package client

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
)

type SnoozeRuleWithID struct {
	SnoozeRule
	ID string `json:"id"`
}

type SnoozeRule struct {
	Title    string   `json:"title"`
	Scope    Scope    `json:"scope"`
	Schedule Schedule `json:"schedule"`
}

type Scope struct {
	Basic *BasicTargeting `json:"basic,omitempty"`
}

type BasicTargeting struct {
	ScopeFilters []ScopeFilter `json:"scope_filters"`
}

type ScopeFilter struct {
	AlertIDs       []string   `json:"alert_ids"`
	LabelPredicate *Predicate `json:"label_predicate"`
}

type Predicate struct {
	Operator string          `json:"operator"`
	Labels   []ResourceLabel `json:"labels"`
}

type ResourceLabel struct {
	Key   string `json:"key"`
	Value string `json:"value"`
}

type Schedule struct {
	OneTime   *OneTimeSchedule   `json:"one_time,omitempty"`
	Recurring *RecurringSchedule `json:"recurring,omitempty"`
}

type OneTimeSchedule struct {
	Timezone      string `json:"timezone"`
	StartDateTime string `json:"start_date_time"`
	EndDateTime   string `json:"end_date_time,omitempty"`
}

type RecurringSchedule struct {
	Timezone  string         `json:"timezone"`
	StartDate string         `json:"start_date"`
	EndDate   string         `json:"end_date,omitempty"`
	Schedules []Reoccurrence `json:"schedules"`
}

type Reoccurrence struct {
	Name           string  `json:"name"`
	StartTime      string  `json:"start_time"`
	DurationMillis int64   `json:"duration_millis"`
	Cadence        Cadence `json:"cadence"`
}

type Cadence struct {
	DaysOfWeek string `json:"days_of_week"`
}

func getSnoozeRuleURL(project string, id string) string {
	base := fmt.Sprintf("projects/%v/snooze_rules", project)
	if id != "" {
		return fmt.Sprintf("%v/%v", base, id)
	}
	return base
}

func getSnoozeRuleValidateURL(project string) string {
	return fmt.Sprintf("projects/%v/snooze_rules_validate", project)
}

func (c *Client) CreateSnoozeRule(
	ctx context.Context,
	projectName string,
	snoozeRule SnoozeRule) (SnoozeRuleWithID, error) {

	var (
		respRule SnoozeRuleWithID
		resp     Envelope
	)

	bytes, err := json.Marshal(snoozeRule)
	if err != nil {
		return respRule, err
	}

	url := getSnoozeRuleURL(projectName, "")

	err = c.CallAPI(ctx, "POST", url, Envelope{Data: bytes}, &resp)
	if err != nil {
		return respRule, err
	}

	err = json.Unmarshal(resp.Data, &respRule)
	if err != nil {
		return respRule, err
	}
	return respRule, err
}

func (c *Client) UpdateSnoozeRule(
	ctx context.Context,
	projectName string,
	id string,
	snoozeRule SnoozeRule,
) (SnoozeRuleWithID, error) {
	var (
		respRule SnoozeRuleWithID
		resp     Envelope
	)

	bytes, err := json.Marshal(snoozeRule)
	if err != nil {
		return respRule, err
	}

	url := getSnoozeRuleURL(projectName, id)

	err = c.CallAPI(ctx, "PUT", url, Envelope{Data: bytes}, &resp)
	if err != nil {
		return respRule, err
	}

	err = json.Unmarshal(resp.Data, &respRule)
	if err != nil {
		return respRule, err
	}

	return respRule, err
}

func (c *Client) GetSnoozeRule(ctx context.Context, projectName string, conditionID string) (*SnoozeRuleWithID, error) {
	var (
		respRule SnoozeRuleWithID
		resp     Envelope
	)

	url := getSnoozeRuleURL(projectName, conditionID)
	err := c.CallAPI(ctx, "GET", url, nil, &resp)
	if err != nil {
		return nil, err
	}

	err = json.Unmarshal(resp.Data, &respRule)
	if err != nil {
		return nil, err
	}
	return &respRule, err
}

func (c *Client) DeleteSnoozeRule(ctx context.Context, projectName string, conditionID string) error {
	url := getSnoozeRuleURL(projectName, conditionID)

	err := c.CallAPI(ctx, "DELETE", url, nil, nil)
	if err != nil {
		apiClientError, ok := err.(APIResponseCarrier)
		if !ok || apiClientError.GetStatusCode() != http.StatusNoContent {
			return err
		}
	}
	return nil
}

func (c *Client) ValidateSnoozeRule(ctx context.Context, projectName string, snoozeRule SnoozeRule) (bool, string, error) {
	var (
		resp Envelope
	)

	bytes, err := json.Marshal(snoozeRule)
	if err != nil {
		return false, "", err
	}

	url := getSnoozeRuleValidateURL(projectName)

	err = c.CallAPI(ctx, "POST", url, Envelope{Data: bytes}, &resp)
	if err != nil {
		return false, "", err
	}

	type ValidationResponse struct {
		IsValid         bool   `json:"is_valid"`
		ValidationError string `json:"validation_error"`
	}
	var validationResponse ValidationResponse

	err = json.Unmarshal(resp.Data, &validationResponse)
	if err != nil {
		return false, "", err
	}

	return validationResponse.IsValid, validationResponse.ValidationError, nil
}
