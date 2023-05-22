package client

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
)

type UnifiedDashboard struct {
	Type       string                     `json:"type"`
	ID         string                     `json:"id"`
	Attributes UnifiedDashboardAttributes `json:"attributes,omitempty"`
}

type UnifiedDashboardAttributes struct {
	Name              string             `json:"name"`
	Description       string             `json:"description"`
	Charts            []UnifiedChart     `json:"charts"`
	Groups            []UnifiedGroup     `json:"groups"`
	Labels            []Label            `json:"labels"`
	TemplateVariables []TemplateVariable `json:"template_variables"`
}

type UnifiedGroup struct {
	ID             string         `json:"id"`
	Rank           int            `json:"rank"`
	Title          string         `json:"title"`
	VisibilityType string         `json:"visibility_type"`
	Charts         []UnifiedChart `json:"charts"`
	Labels         []Label        `json:"labels"`
}

type UnifiedPosition struct {
	XPos   int `json:"x-pos"`
	YPos   int `json:"y-pos"`
	Width  int `json:"width"`
	Height int `json:"height"`
}

type UnifiedChart struct {
	Rank          int                         `json:"rank"`
	Position      UnifiedPosition             `json:"position"`
	ID            string                      `json:"id"`
	Title         string                      `json:"title"`
	ChartType     string                      `json:"chart-type"`
	YAxis         *YAxis                      `json:"y-axis"`
	MetricQueries []MetricQueryWithAttributes `json:"metric-queries"`
	Text          string                      `json:"text"`
}

type Label struct {
	Key   string `json:"label_key"`
	Value string `json:"label_value"`
}

type YAxis struct {
	Min float64 `json:"min"`
	Max float64 `json:"max"`
}

type MetricGroupBy struct {
	LabelKeys         []string `json:"label-keys"`
	AggregationMethod string   `json:"aggregation-method"`
}

type TemplateVariable struct {
	Name                   string   `json:"name"`
	DefaultValues          []string `json:"default_values"`
	SuggestionAttributeKey string   `json:"suggestion_attribute_key"`
}

func getUnifiedDashboardURL(project, id string, query map[string]string) string {
	path := fmt.Sprintf(
		"projects/%s/metric_dashboards",
		url.PathEscape(project),
	)
	if id != "" {
		path += "/" + url.PathEscape(id)
	}
	u := url.URL{Path: path}
	if len(query) > 0 {
		q := u.Query()
		for k, v := range query {
			q.Set(k, v)
		}
		u.RawQuery = q.Encode()
	}
	return u.String()
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
			Name:              dashboard.Attributes.Name,
			Description:       dashboard.Attributes.Description,
			Groups:            dashboard.Attributes.Groups,
			Labels:            dashboard.Attributes.Labels,
			TemplateVariables: dashboard.Attributes.TemplateVariables,
		},
	})

	if err != nil {
		return cond, err
	}

	url := getUnifiedDashboardURL(projectName, "", nil)

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

func (c *Client) GetUnifiedDashboard(ctx context.Context, projectName string, id string, convertToQueryString bool) (*UnifiedDashboard, error) {
	var (
		d    *UnifiedDashboard
		resp Envelope
	)

	var q map[string]string
	if convertToQueryString {
		q = map[string]string{"query_format": "query_string"}
	}

	url := getUnifiedDashboardURL(projectName, id, q)
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

	url := getUnifiedDashboardURL(projectName, dashboardID, nil)
	err = c.CallAPI(ctx, "PUT", url, Envelope{Data: bytes}, &resp)
	if err != nil {
		return d, err
	}

	err = json.Unmarshal(resp.Data, &d)
	return d, err
}

func (c *Client) DeleteUnifiedDashboard(ctx context.Context, projectName string, dashboardID string) error {
	url := getUnifiedDashboardURL(projectName, dashboardID, nil)

	err := c.CallAPI(ctx, "DELETE", url, nil, nil)
	if err != nil {
		apiClientError, ok := err.(APIResponseCarrier)
		if !ok || apiClientError.GetStatusCode() != http.StatusNoContent {
			return err
		}
	}
	return nil
}
