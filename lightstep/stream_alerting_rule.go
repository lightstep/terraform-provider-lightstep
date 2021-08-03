package lightstep

import (
	"encoding/json"
	"fmt"
	"net/http"
)

type StreamAlertingRule struct {
	Type          string                          `json:"type,omitempty"`
	ID            string                          `json:"id,omitempty"`
	Attributes    StreamAlertingAttributes        `json:"attributes"`
	Relationships StreamAlertingRuleRelationships `json:"relationships,omitempty"`
}

type StreamAlertingAttributes struct {
	UpdateInterval int `json:"update-interval-ms"`
}

type StreamAlertingRuleRelationships struct {
	Condition   ConditionStream `json:"condition"`
	Destination Destination     `json:"destination"`
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

	bytes, err := json.Marshal(StreamAlertingRule{
		Type: "alerting_rule",
		Attributes: StreamAlertingAttributes{
			UpdateInterval: updateInterval,
		},
		Relationships: StreamAlertingRuleRelationships{
			Condition: ConditionStream{
				ID:   conditionID,
				Type: "condition",
			},
			Destination: Destination{
				Type: "destination",
				ID:   destinationID,
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

// DeleteAlertingRule delete an alerting rule. Not currently implemented.
// https://lightstep.atlassian.net/browse/LS-24956
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
