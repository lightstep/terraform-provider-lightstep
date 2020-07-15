package lightstep

import (
	"encoding/json"
	"fmt"
	"net/http"
)

type Condition struct {
	Type          string                 `json:"type,omitempty"`
	ID            string                 `json:"id,omitempty"`
	Attributes    ConditionAttributes    `json:"attributes"`
	Relationships ConditionRelationships `json:"relationships,omitempty"`
}

type ConditionAttributes struct {
	Name               string                 `json:"name,omitempty"`
	EvaluationWindowMS int                    `json:"eval-window-ms,omitempty"`
	Expression         string                 `json:"expression,omitempty"`
	CustomData         map[string]interface{} `json:"custom-data,omitempty"`
}

type ConditionRelationships struct {
	Stream ConditionStream `json:"stream"`
}

type ConditionStream struct {
	ID   string `json:"id"`
	Type string `json:"type"`
}

func (c *Client) CreateCondition(
	projectName string,
	conditionName string,
	expression string,
	evaluationWindowMS int,
	streamID string) (Condition, error) {

	var (
		cond Condition
		resp Envelope
	)

	bytes, err := json.Marshal(Condition{
		Type: "condition",
		Attributes: ConditionAttributes{
			Name:               conditionName,
			EvaluationWindowMS: evaluationWindowMS,
			Expression:         expression,
			CustomData:         nil,
		},
		Relationships: ConditionRelationships{
			Stream: ConditionStream{
				ID:   streamID,
				Type: "stream",
			},
		},
	})
	if err != nil {
		return cond, err
	}

	err = c.CallAPI("POST", fmt.Sprintf("projects/%v/conditions", projectName), Envelope{Data: bytes}, &resp)
	if err != nil {
		return cond, err
	}

	err = json.Unmarshal(resp.Data, &cond)
	if err != nil {
		return cond, err
	}
	return cond, err
}

func (c *Client) UpdateCondition(
	projectName string,
	conditionID string,
	attributes ConditionAttributes,
) (Condition, error) {
	var (
		cond Condition
		resp Envelope
	)

	bytes, err := json.Marshal(&Condition{
		ID:         conditionID,
		Attributes: attributes,
	})
	if err != nil {
		return cond, err
	}

	err = c.CallAPI(
		"PATCH",
		fmt.Sprintf("projects/%v/conditions/%v", projectName, conditionID),
		Envelope{Data: bytes},
		&resp)
	if err != nil {
		return cond, err
	}

	err = json.Unmarshal(resp.Data, &cond)
	if err != nil {
		return cond, err
	}

	return cond, err
}

func (c *Client) GetCondition(projectName string, conditionID string) (Condition, error) {
	var (
		cond Condition
		resp Envelope
	)

	err := c.CallAPI("GET", fmt.Sprintf("projects/%v/conditions/%v", projectName, conditionID), nil, &resp)
	if err != nil {
		return cond, err
	}

	err = json.Unmarshal(resp.Data, &cond)
	if err != nil {
		return cond, err
	}
	return cond, err
}

func (c *Client) DeleteCondition(projectName string, conditionID string) error {
	err := c.CallAPI("DELETE", fmt.Sprintf("projects/%v/conditions/%v", projectName, conditionID), nil, nil)
	if err != nil {
		apiClientError := err.(APIResponseCarrier)
		if apiClientError.GetHTTPResponse().StatusCode != http.StatusNoContent {
			return err
		}
	}
	return nil
}
