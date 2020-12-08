package lightstep

import (
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
	Name          string `json:"name"`
	Type          string `json:"condition_type"`
	Expression    `json:"expression"`
	Queries       []MetricQueryWithAttributes `json:"metric-queries"`
	AlertingRules []AlertingRule              `json:"alerting-rules,omitempty"`
}

type AlertingRule struct {
	MessageDestinationID string  `json:"message-destination-client-id"`
	UpdateInterval       int     `json:"update-interval-ms"`
	MatchOn              MatchOn `json:"match-on,omitempty"`
}

type Expression struct {
	Thresholds         `json:"thresholds"`
	Operand            string `json:"operand"`
	EvaluationWindow   int    `json:"evaluation-window-ms"`
	EvaluationCriteria string `json:"evaluation-criteria"`
	IsMulti            bool   `json:"is-multi-alert,omitempty"`
	IsNoData           bool   `json:"enable-no-data-alert,omitempty"`
}

type Thresholds struct {
	Warning  int `json:"warning,omitempty"`
	Critical int `json:"critical"`
}

type MetricQueryWithAttributes struct {
	Name    string      `json:"query-name"`
	Type    string      `json:"query-type"`
	Hidden  bool        `json:"hidden"`
	Display string      `json:"display-type"`
	Query   MetricQuery `json:"metric-query"`
}

type MetricQuery struct {
	Metric             string        `json:"metric"`
	Filters            []LabelFilter `json:"filters,omitempty"`
	TimeseriesOperator string        `json:"timeseries-operator"`
	GroupBy            GroupBy       `json:"group-by,omitempty"`
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
		},
	})

	if err != nil {
		return cond, err
	}

	url := getURL(projectName, "")

	err = c.CallAPI("POST", url, Envelope{Data: bytes}, &resp)
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

	err = c.CallAPI("PUT", url, Envelope{Data: bytes}, &resp)
	if err != nil {
		return cond, err
	}

	err = json.Unmarshal(resp.Data, &cond)
	if err != nil {
		return cond, err
	}

	return cond, err
}

func (c *Client) GetMetricCondition(projectName string, conditionID string) (MetricCondition, error) {
	var (
		cond MetricCondition
		resp Envelope
	)

	url := getURL(projectName, conditionID)
	err := c.CallAPI("GET", url, nil, &resp)
	if err != nil {
		return cond, err
	}

	err = json.Unmarshal(resp.Data, &cond)
	if err != nil {
		return cond, err
	}
	return cond, err
}

func (c *Client) DeleteMetricCondition(projectName string, conditionID string) error {
	url := getURL(projectName, conditionID)

	err := c.CallAPI("DELETE", url, nil, nil)
	if err != nil {
		apiClientError := err.(APIResponseCarrier)
		if apiClientError.GetHTTPResponse().StatusCode != http.StatusNoContent {
			return err
		}
	}
	return nil
}
