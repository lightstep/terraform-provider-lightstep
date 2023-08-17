package client

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
)

const InferredServiceRuleType = "inferred_service_rule"

type InferredServiceRuleRequest struct {
	CreateRequest
	Attributes InferredServiceRuleRequestAttributes `json:"attributes,omitempty"`
}

type InferredServiceRuleRequestAttributes struct {
	Name             string            `json:"name"`
	Description      *string           `json:"description"`
	AttributeFilters []AttributeFilter `json:"attribute-filters"`
	GroupByKeys      []string          `json:"group-by-keys"`
}

type InferredServiceRuleResponse struct {
	CreateResponse
	Attributes InferredServiceRuleResponseAttributes `json:"attributes,omitempty"`
}

type InferredServiceRuleResponseAttributes struct {
	Name             string            `json:"name"`
	Description      string            `json:"description,omitempty"`
	AttributeFilters []AttributeFilter `json:"attribute-filters"`
	GroupByKeys      []string          `json:"group-by-keys,omitempty"`
}

type AttributeFilter struct {
	Key    string   `json:"key"`
	Values []string `json:"values"`
}

func (c *Client) CreateInferredServiceRule(
	ctx context.Context,
	projectName string,
	requestAttributes InferredServiceRuleRequestAttributes,
) (InferredServiceRuleResponse, error) {
	var (
		inferredServiceRuleResponse InferredServiceRuleResponse
		apiResponse                 Envelope
	)

	request := InferredServiceRuleRequest{
		CreateRequest: CreateRequest{Type: InferredServiceRuleType},
		Attributes:    requestAttributes,
	}
	bytes, err := json.Marshal(request)
	if err != nil {
		return inferredServiceRuleResponse, err
	}

	apiPath := getInferredServiceRuleUrl(projectName)
	err = c.CallAPI(ctx, "POST", apiPath, Envelope{Data: bytes}, &apiResponse)
	if err != nil {
		return inferredServiceRuleResponse, err
	}

	err = json.Unmarshal(apiResponse.Data, &inferredServiceRuleResponse)

	return inferredServiceRuleResponse, err
}

func (c *Client) GetInferredServiceRule(
	ctx context.Context,
	project string,
	inferredServiceRuleId string,
) (InferredServiceRuleResponse, error) {
	var (
		inferredServiceRuleResponse InferredServiceRuleResponse
		apiResponse                 Envelope
	)

	apiPath := getInferredServiceRuleUrlWithId(project, inferredServiceRuleId)
	err := c.CallAPI(ctx, "GET", apiPath, nil, &apiResponse)
	if err != nil {
		return inferredServiceRuleResponse, err
	}

	err = json.Unmarshal(apiResponse.Data, &inferredServiceRuleResponse)
	return inferredServiceRuleResponse, err
}

func (c *Client) UpdateInferredServiceRule(
	ctx context.Context,
	projectName string,
	inferredServiceRuleID string,
	requestAttributes InferredServiceRuleRequestAttributes,
) (InferredServiceRuleResponse, error) {
	var (
		inferredServiceRuleResponse InferredServiceRuleResponse
		response                    Envelope
	)

	request := InferredServiceRuleRequest{
		CreateRequest: CreateRequest{Type: InferredServiceRuleType},
		Attributes:    requestAttributes,
	}
	bytes, err := json.Marshal(request)
	if err != nil {
		return inferredServiceRuleResponse, err
	}

	apiPath := getInferredServiceRuleUrlWithId(projectName, inferredServiceRuleID)
	err = c.CallAPI(ctx, "PUT", apiPath, Envelope{Data: bytes}, &response)
	if err != nil {
		return inferredServiceRuleResponse, err
	}

	err = json.Unmarshal(response.Data, &inferredServiceRuleResponse)
	return inferredServiceRuleResponse, err
}

func (c *Client) DeleteInferredServiceRule(
	ctx context.Context,
	projectName string,
	inferredServiceRuleID string,
) error {
	apiPath := getInferredServiceRuleUrlWithId(projectName, inferredServiceRuleID)
	err := c.CallAPI(ctx, "DELETE", apiPath, nil, nil)
	if err != nil {
		apiClientError, ok := err.(APIResponseCarrier)
		if !ok || apiClientError.GetStatusCode() != http.StatusNoContent {
			return err
		}
	}

	return nil
}

func getInferredServiceRuleUrl(project string) string {
	const inferredServiceRuleBasePathTemplate = "projects/%s/inferred_service_rules"

	path := fmt.Sprintf(
		inferredServiceRuleBasePathTemplate,
		url.PathEscape(project),
	)
	asUrl := url.URL{Path: path}
	return asUrl.String()
}

func getInferredServiceRuleUrlWithId(project, id string) string {
	path := getInferredServiceRuleUrl(project)

	if id != "" {
		path += "/" + url.PathEscape(id)
	}

	asUrl := url.URL{Path: path}
	return asUrl.String()
}
