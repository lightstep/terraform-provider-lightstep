package lightstep

import (
	"encoding/json"
	"fmt"
	"net/http"
)

type CreateRequest struct {
	Type string `json:"type"`
}

type Response struct {
	Type string `json:"type"`
	ID   string `json:"id"`
}

type RelatedResourceObject struct {
	Type string `json:"type"`
	ID   string `json:"id"`
}

type RelatedResourceWithLinks struct {
	Links Links                 `json:"links"`
	Data  RelatedResourceObject `json:"data"`
}

type StreamAlertingRuleResponse struct {
	Response
	Attributes    StreamAlertingRuleAttributes            `json:"attributes,omitempty"`
	Relationships StreamAlertingRuleResponseRelationships `json:"relationships,omitempty"`
	Links         Links                                   `json:"links"`
}

type StreamAlertingRuleRequest struct {
	CreateRequest
	Attributes    StreamAlertingRuleAttributes           `json:"attributes,omitempty"`
	Relationships StreamAlertingRuleRequestRelationships `json:"relationships,omitempty"`
}

type StreamAlertingRuleAttributes struct {
	UpdateInterval int `json:"update-interval-ms"`
}

type StreamAlertingRuleRequestRelationships struct {
	Condition   RelatedResourceObject `json:"condition"`
	Destination RelatedResourceObject `json:"destination"`
}

type StreamAlertingRuleResponseRelationships struct {
	Condition   RelatedResourceWithLinks `json:"condition"`
	Destination RelatedResourceWithLinks `json:"destination"`
	Project     RelatedResourceWithLinks `json:"project"`
	Stream      RelatedResourceWithLinks `json:"stream"`
}

func (c *Client) CreateAlertingRule(
	projectName string,
	updateInterval int,
	destinationID string,
	conditionID string) (StreamAlertingRuleResponse, error) {

	var (
		rule StreamAlertingRuleResponse
		resp Envelope
	)

	bytes, err := json.Marshal(StreamAlertingRuleRequest{
		CreateRequest: CreateRequest{
			Type: "alerting_rule",
		},
		Attributes: StreamAlertingRuleAttributes{
			UpdateInterval: updateInterval,
		},
		Relationships: StreamAlertingRuleRequestRelationships{
			Condition: RelatedResourceObject{
				ID:   conditionID,
				Type: "condition",
			},
			Destination: RelatedResourceObject{
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

func (c *Client) GetAlertingRule(projectName string, alertingRuleID string) (*StreamAlertingRuleResponse, error) {
	var (
		rule StreamAlertingRuleResponse
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
