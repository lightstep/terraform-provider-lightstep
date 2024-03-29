package client

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
)

type StreamCondition struct {
	Type          string                       `json:"type,omitempty"`
	ID            string                       `json:"id,omitempty"`
	Attributes    StreamConditionAttributes    `json:"attributes"`
	Relationships StreamConditionRelationships `json:"relationships,omitempty"`
}

type StreamConditionAttributes struct {
	Name               string                 `json:"name,omitempty"`
	EvaluationWindowMS int                    `json:"eval-window-ms,omitempty"`
	Expression         string                 `json:"expression,omitempty"`
	CustomData         map[string]interface{} `json:"custom_data,omitempty"`
}

type StreamConditionRelationships struct {
	Stream ConditionStream `json:"stream"`
}

type ConditionStream struct {
	ID    string `json:"id"`
	Type  string `json:"type"`
	Links Links  `json:"links"`
}

type Links struct {
	Related string `json:"related"`
	Self    string `json:"self"`
}

func (c *Client) CreateStreamCondition(
	ctx context.Context,
	projectName string,
	conditionName string,
	expression string,
	evaluationWindowMS int,
	streamID string) (StreamCondition, error) {

	var (
		cond StreamCondition
		resp Envelope
	)

	bytes, err := json.Marshal(StreamCondition{
		Type: "condition",
		Attributes: StreamConditionAttributes{
			Name:               conditionName,
			EvaluationWindowMS: evaluationWindowMS,
			Expression:         expression,
			CustomData:         nil,
		},
		Relationships: StreamConditionRelationships{
			Stream: ConditionStream{
				ID:   streamID,
				Type: "stream",
			},
		},
	})
	if err != nil {
		return cond, err
	}

	err = c.CallAPI(ctx, "POST", fmt.Sprintf("projects/%v/conditions", projectName), Envelope{Data: bytes}, &resp)
	if err != nil {
		return cond, err
	}

	err = json.Unmarshal(resp.Data, &cond)
	if err != nil {
		return cond, err
	}

	return cond, nil
}

func (c *Client) UpdateStreamCondition(
	ctx context.Context,
	projectName string,
	conditionID string,
	attributes StreamConditionAttributes,
) (*StreamCondition, error) {
	var (
		cond *StreamCondition
		resp Envelope
	)

	bytes, err := json.Marshal(&StreamCondition{
		ID:         conditionID,
		Attributes: attributes,
	})
	if err != nil {
		return cond, err
	}

	err = c.CallAPI(
		ctx,
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

func (c *Client) GetStreamCondition(ctx context.Context, projectName string, conditionID string) (*StreamCondition, error) {
	var (
		cond StreamCondition
		resp Envelope
	)
	err := c.CallAPI(ctx, "GET", fmt.Sprintf("projects/%v/conditions/%v", projectName, conditionID), nil, &resp)
	if err != nil {
		return nil, err
	}

	err = json.Unmarshal(resp.Data, &cond)
	if err != nil {
		return nil, err
	}
	return &cond, err
}

func (c *Client) DeleteStreamCondition(ctx context.Context, projectName string, conditionID string) error {
	err := c.CallAPI(ctx, "DELETE", fmt.Sprintf("projects/%v/conditions/%v", projectName, conditionID), nil, nil)
	if err != nil {
		apiClientError, ok := err.(APIResponseCarrier)
		if !ok || apiClientError.GetStatusCode() != http.StatusNoContent {
			return err
		}
	}
	return nil
}
