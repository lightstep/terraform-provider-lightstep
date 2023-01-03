package client

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
)

type MetricCondition struct {
	ID         string                    `json:"id"`
	Type       string                    `json:"type"`
	Attributes MetricConditionAttributes `json:"attributes"`
}

type MetricConditionAttributes struct {
	Name          string                      `json:"name"`
	Description   string                      `json:"description"`
	Type          string                      `json:"condition_type"`
	Expression    Expression                  `json:"expression"`
	Queries       []MetricQueryWithAttributes `json:"metric-queries"`
	AlertingRules []AlertingRule              `json:"alerting-rules,omitempty"`
}

type AlertingRule struct {
	MessageDestinationID string  `json:"message-destination-client-id"`
	UpdateInterval       int     `json:"update-interval-ms"`
	MatchOn              MatchOn `json:"match-on,omitempty"`
}

type Expression struct {
	Thresholds Thresholds `json:"thresholds"`
	Operand    string     `json:"operand"`
	IsMulti    bool       `json:"is-multi-alert,omitempty"`
	IsNoData   bool       `json:"enable-no-data-alert,omitempty"`
}

type Thresholds struct {
	Warning  *float64 `json:"warning,omitempty"`
	Critical *float64 `json:"critical,omitempty"`
}

type MetricQueryWithAttributes struct {
	Name           string         `json:"query-name"`
	Type           string         `json:"query-type"`
	Hidden         bool           `json:"hidden"`
	Display        string         `json:"display-type"`
	Query          MetricQuery    `json:"metric-query"`
	SpansQuery     SpansQuery     `json:"spans-query,omitempty"`
	CompositeQuery CompositeQuery `json:"composite-query,omitempty"`
	TQLQuery       string         `json:"tql-query"`
}

type MetricQuery struct {
	Metric                          string                `json:"metric"`
	Filters                         []LabelFilter         `json:"filters,omitempty"`
	TimeseriesOperator              string                `json:"timeseries-operator"`
	TimeseriesOperatorInputWindowMs *int                  `json:"timeseries-operator-input-window-ms,omitempty"`
	GroupBy                         GroupBy               `json:"group-by,omitempty"`
	FinalWindowOperation            *FinalWindowOperation `json:"final-window-operation,omitempty"`
}

type FinalWindowOperation struct {
	Operator      string `json:"operator"`
	InputWindowMs int    `json:"input-window-ms"`
}

type SpansQuery struct {
	Query                 string                `json:"query"`
	Operator              string                `json:"operator"`
	OperatorInputWindowMs *int                  `json:"operator-input-window-ms,omitempty"`
	LatencyPercentiles    []float64             `json:"latency-percentiles,omitempty"`
	GroupByKeys           []string              `json:"group-by,omitempty"`
	FinalWindowOperation  *FinalWindowOperation `json:"final-window-operation,omitempty"`
}

type CompositeQuery struct {
	FinalWindowOperation *FinalWindowOperation `json:"final-window-operation"`
}

type LabelFilter struct {
	Key     string `json:"key"`
	Value   string `json:"value"`
	Operand string `json:"operand"`
}

type GroupBy struct {
	LabelKeys   []string `json:"label-keys"`
	Aggregation string   `json:"aggregation-method"`
}

type MatchOn struct {
	GroupBy []LabelFilter `json:"group-by"`
}

func getURL(project string, id string) string {
	base := fmt.Sprintf("projects/%v/metric_alerts", project)
	if id != "" {
		return fmt.Sprintf("%v/%v", base, id)
	}
	return base
}

func (c *Client) CreateMetricCondition(
	ctx context.Context,
	projectName string,
	condition MetricCondition) (MetricCondition, error) {

	var (
		cond MetricCondition
		resp Envelope
	)

	bytes, err := json.Marshal(MetricCondition{
		Type: condition.Type,
		Attributes: MetricConditionAttributes{
			Name:          condition.Attributes.Name,
			Type:          condition.Type,
			Expression:    condition.Attributes.Expression,
			Queries:       condition.Attributes.Queries,
			AlertingRules: condition.Attributes.AlertingRules,
			Description:   condition.Attributes.Description,
		},
	})

	if err != nil {
		return cond, err
	}

	url := getURL(projectName, "")

	err = c.CallAPI(ctx, "POST", url, Envelope{Data: bytes}, &resp)
	if err != nil {
		return cond, err
	}

	err = json.Unmarshal(resp.Data, &cond)
	if err != nil {
		return cond, err
	}
	return cond, err
}

func (c *Client) UpdateMetricCondition(
	ctx context.Context,
	projectName string,
	conditionID string,
	attributes MetricConditionAttributes,
) (MetricCondition, error) {
	var (
		cond MetricCondition
		resp Envelope
	)

	bytes, err := json.Marshal(&MetricCondition{
		Type:       "metric_alert",
		ID:         conditionID,
		Attributes: attributes,
	})
	if err != nil {
		return cond, err
	}

	url := getURL(projectName, conditionID)

	err = c.CallAPI(ctx, "PUT", url, Envelope{Data: bytes}, &resp)
	if err != nil {
		return cond, err
	}

	err = json.Unmarshal(resp.Data, &cond)
	if err != nil {
		return cond, err
	}

	return cond, err
}

func (c *Client) GetMetricCondition(ctx context.Context, projectName string, conditionID string) (*MetricCondition, error) {
	var (
		cond MetricCondition
		resp Envelope
	)

	url := getURL(projectName, conditionID)
	err := c.CallAPI(ctx, "GET", url, nil, &resp)
	if err != nil {
		return nil, err
	}

	err = json.Unmarshal(resp.Data, &cond)
	if err != nil {
		return nil, err
	}
	return &cond, err
}

func (c *Client) DeleteMetricCondition(ctx context.Context, projectName string, conditionID string) error {
	url := getURL(projectName, conditionID)

	err := c.CallAPI(ctx, "DELETE", url, nil, nil)
	if err != nil {
		apiClientError := err.(APIResponseCarrier)
		if apiClientError.GetHTTPResponse().StatusCode != http.StatusNoContent {
			return err
		}
	}
	return nil
}
