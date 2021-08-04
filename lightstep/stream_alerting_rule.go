package lightstep

import (
	"encoding/json"
	"fmt"
	"net/http"
)

type StreamAlertingRule struct {
	Type          string                          `json:"type,omitempty"`
	ID            string                          `json:"id,omitempty"`
	Attributes    StreamAlertingAttributes        `json:"attributes,omitempty"`
	Relationships StreamAlertingRuleRelationships `json:"relationships,omitempty"`
}

type StreamAlertingAttributes struct {
	UpdateInterval int `json:"update-interval-ms"`
}

type StreamAlertingRuleRelationships struct {
	Condition   StreamAlertingRuleCondition   `json:"condition"`
	Destination StreamAlertingRuleDestination `json:"destination"`
}

type StreamAlertingRuleCondition struct {
	Links Links                           `json:"links"`
	Data  StreamAlertingRuleConditionData `json:"data"`
}

type StreamAlertingRuleConditionData struct {
	Type string `json:"type"`
	ID   string `json:"id"`
}

type StreamAlertingRuleDestination struct {
	Links Links                           `json:"links"`
	Data  StreamAlertingRuleConditionData `json:"data"`
}

// The request and response json is different!
type StreamAlertingRuleRequest struct {
	Type          string                                 `json:"type,omitempty"`
	ID            string                                 `json:"id,omitempty"`
	Attributes    StreamAlertingAttributes               `json:"attributes,omitempty"`
	Relationships StreamAlertingRuleRequestRelationships `json:"relationships,omitempty"`
}

type StreamAlertingRuleRequestRelationships struct {
	Condition   StreamAlertingRuleRequestRelationshipsDesc `json:"condition"`
	Destination StreamAlertingRuleRequestRelationshipsDesc `json:"destination"`
}

type StreamAlertingRuleRequestRelationshipsDesc struct {
	ID   string `json:"id"`
	Type string `json:"type"`
}

func (c *Client) CreateAlertingRule(
	projectName string,
	updateInterval int,
	destinationID string,
	conditionID string) (StreamAlertingRule, error) {

	var (
		rule StreamAlertingRule
		resp Envelope
	)

	bytes, err := json.Marshal(StreamAlertingRuleRequest{
		Type: "alerting_rule",
		Attributes: StreamAlertingAttributes{
			UpdateInterval: updateInterval,
		},
		Relationships: StreamAlertingRuleRequestRelationships{
			Condition: StreamAlertingRuleRequestRelationshipsDesc{
				ID:   conditionID,
				Type: "condition",
			},
			Destination: StreamAlertingRuleRequestRelationshipsDesc{
				ID:   destinationID,
				Type: "destination",
			},
		},
	})
	if err != nil {
		return rule, err
	}

	err = c.CallAPI("POST", fmt.Sprintf("projects/%v/alerting_rules", projectName), Envelope{Data: bytes}, &resp)
	if err != nil {
		return rule, err
	}

	err = json.Unmarshal(resp.Data, &rule)
	if err != nil {
		return rule, err
	}
	return rule, err
}

func (c *Client) GetAlertingRule(projectName string, alertingRuleID string) (*StreamAlertingRule, error) {
	var (
		rule StreamAlertingRule
		resp Envelope
	)
	err := c.CallAPI("GET", fmt.Sprintf("projects/%v/alerting_rules/%v", projectName, alertingRuleID), nil, &resp)
	if err != nil {
		return nil, err
	}

	err = json.Unmarshal(resp.Data, &rule)
	if err != nil {
		return nil, err
	}
	return &rule, err
}

func (c *Client) DeleteAlertingRule(projectName string, alertingRuleID string) error {
	err := c.CallAPI("DELETE", fmt.Sprintf("projects/%v/alerting_rules/%v", projectName, alertingRuleID), nil, nil)
	if err != nil {
		apiClientError := err.(APIResponseCarrier)
		if apiClientError.GetHTTPResponse().StatusCode != http.StatusNoContent {
			return err
		}
	}
	return nil
}
