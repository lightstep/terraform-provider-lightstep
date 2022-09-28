package client

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
)

type UnifiedDashboard struct {
	Type       string                     `json:"type"`
	ID         string                     `json:"id"`
	Attributes UnifiedDashboardAttributes `json:"attributes,omitempty"`
}

type UnifiedDashboardAttributes struct {
	Name   string         `json:"name"`
	Charts []UnifiedChart `json:"charts"`
}

type UnifiedChart struct {
	Rank          int                         `json:"rank"`
	ID            string                      `json:"id"`
	Title         string                      `json:"title"`
	ChartType     string                      `json:"chart-type"`
	YAxis         *YAxis                      `json:"y-axis"`
	MetricQueries []MetricQueryWithAttributes `json:"metric-queries"`
}

type YAxis struct {
	Min float64 `json:"min"`
	Max float64 `json:"max"`
}

type MetricGroupBy struct {
	LabelKeys         []string `json:"label-keys"`
	AggregationMethod string   `json:"aggregation-method"`
}

func getUnifiedDashboardURL(project string, id string) string {
	base := fmt.Sprintf("projects/%s/metric_dashboards", project)

	if id != "" {
		return fmt.Sprintf("%s/%s", base, id)
	}
	return base
}

func (c *Client) CreateUnifiedDashboard(
	ctx context.Context,
	projectName string,
	dashboard UnifiedDashboard) (UnifiedDashboard, error) {

	var (
		cond UnifiedDashboard
		resp Envelope
	)

	bytes, err := json.Marshal(UnifiedDashboard{
		Type: dashboard.Type,
		Attributes: UnifiedDashboardAttributes{
			Name:   dashboard.Attributes.Name,
			Charts: dashboard.Attributes.Charts,
		},
	})

	if err != nil {
		return cond, err
	}

	url := getUnifiedDashboardURL(projectName, "")

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

func (c *Client) GetUnifiedDashboard(ctx context.Context, projectName string, id string) (*UnifiedDashboard, error) {
	var (
		d    *UnifiedDashboard
		resp Envelope
	)

	url := getUnifiedDashboardURL(projectName, id)
	err := c.CallAPI(ctx, "GET", url, nil, &resp)
	if err != nil {
		return nil, err
	}

	err = json.Unmarshal(resp.Data, &d)
	return d, err
}

func (c *Client) UpdateUnifiedDashboard(
	ctx context.Context,
	projectName string,
	dashboardID string,
	attributes UnifiedDashboardAttributes,
) (*UnifiedDashboard, error) {
	var (
		d    *UnifiedDashboard
		resp Envelope
	)

	bytes, err := json.Marshal(&UnifiedDashboard{
		Type:       "dashboard",
		ID:         dashboardID,
		Attributes: attributes,
	})
	if err != nil {
		return nil, err
	}

	url := getUnifiedDashboardURL(projectName, dashboardID)
	err = c.CallAPI(ctx, "PUT", url, Envelope{Data: bytes}, &resp)
	if err != nil {
		return d, err
	}

	err = json.Unmarshal(resp.Data, &d)
	return d, err
}

func (c *Client) DeleteUnifiedDashboard(ctx context.Context, projectName string, dashboardID string) error {
	url := getUnifiedDashboardURL(projectName, dashboardID)

	err := c.CallAPI(ctx, "DELETE", url, nil, nil)
	if err != nil {
		apiClientError := err.(APIResponseCarrier)
		if apiClientError.GetHTTPResponse().StatusCode != http.StatusNoContent {
			return err
		}
	}
	return nil
}
