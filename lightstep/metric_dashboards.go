package lightstep

import (
	"encoding/json"
	"fmt"
	"net/http"
)

type MetricDashboard struct {
	Type       string                    `json:"type"`
	ID         string                    `json:"id"`
	Attributes MetricDashboardAttributes `json:"attributes,omitempty"`
}

type MetricDashboardAttributes struct {
	Name   string        `json:"name"`
	Charts []MetricChart `json:"charts"`
}

type MetricChart struct {
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

func getMetricDashboardURL(project string, id string) string {
	base := fmt.Sprintf("projects/%s/metric_dashboards", project)

	if id != "" {
		return fmt.Sprintf("%s/%s", base, id)
	}
	return base
}

func (c *Client) CreateMetricDashboard(
	projectName string,
	dashboard MetricDashboard) (MetricDashboard, error) {

	var (
		cond MetricDashboard
		resp Envelope
	)

	bytes, err := json.Marshal(MetricDashboard{
		Type: dashboard.Type,
		Attributes: MetricDashboardAttributes{
			Name:   dashboard.Attributes.Name,
			Charts: dashboard.Attributes.Charts,
		},
	})

	if err != nil {
		return cond, err
	}

	url := getMetricDashboardURL(projectName, "")

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

func (c *Client) GetMetricDashboard(projectName string, id string) (MetricDashboard, error) {
	var (
		d    MetricDashboard
		resp Envelope
	)

	url := getMetricDashboardURL(projectName, id)
	err := c.CallAPI("GET", url, nil, &resp)
	if err != nil {
		return d, err
	}

	err = json.Unmarshal(resp.Data, &d)
	return d, err
}

func (c *Client) UpdateMetricDashboard(
	projectName string,
	dashboardID string,
	attributes MetricDashboardAttributes,
) (MetricDashboard, error) {
	var (
		d    MetricDashboard
		resp Envelope
	)

	bytes, err := json.Marshal(&MetricDashboard{
		Type:       "dashboard",
		ID:         dashboardID,
		Attributes: attributes,
	})
	if err != nil {
		return d, err
	}

	url := getMetricDashboardURL(projectName, dashboardID)
	err = c.CallAPI("PUT", url, Envelope{Data: bytes}, &resp)
	if err != nil {
		return d, err
	}

	err = json.Unmarshal(resp.Data, &d)
	return d, err
}

func (c *Client) DeleteMetricDashboard(projectName string, dashboardID string) error {
	url := getMetricDashboardURL(projectName, dashboardID)

	err := c.CallAPI("DELETE", url, nil, nil)
	if err != nil {
		apiClientError := err.(APIResponseCarrier)
		if apiClientError.GetHTTPResponse().StatusCode != http.StatusNoContent {
			return err
		}
	}
	return nil
}
