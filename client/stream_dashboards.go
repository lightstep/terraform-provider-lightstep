package client

import (
	"encoding/json"
	"fmt"
	"net/http"
)

type Dashboard struct {
	Type       string              `json:"type,omitempty"`
	ID         string              `json:"id,omitempty"`
	Attributes DashboardAttributes `json:"attributes,omitempty"`
}

type DashboardAttributes struct {
	Name    string   `json:"name"`
	Streams []Stream `json:"streams"`
}

func (c *Client) CreateDashboard(
	projectName string,
	dashboardName string,
	streams []Stream,
) (Dashboard, error) {

	var (
		d    Dashboard
		resp Envelope
	)

	bytes, err := json.Marshal(
		Dashboard{
			Type: "dashboard",
			Attributes: DashboardAttributes{
				Name:    dashboardName,
				Streams: streams,
			},
		})

	if err != nil {
		return d, err
	}

	err = c.CallAPI("POST", fmt.Sprintf("projects/%v/dashboards", projectName), Envelope{Data: bytes}, &resp)
	if err != nil {
		return d, err
	}

	err = json.Unmarshal(resp.Data, &d)
	if err != nil {
		return d, err
	}

	return d, err
}

func (c *Client) UpdateDashboard(
	projectName string,
	dashboardName string,
	streams []Stream,
	dashboardID string,
) (Dashboard, error) {

	var (
		d    Dashboard
		resp Envelope
	)

	bytes, err := json.Marshal(&Dashboard{
		Type: "dashboard",
		ID:   dashboardID,
		Attributes: DashboardAttributes{
			Name:    dashboardName,
			Streams: streams,
		},
	})
	if err != nil {
		return d, err
	}

	err = c.CallAPI("PATCH", fmt.Sprintf("projects/%v/dashboards/%v", projectName, dashboardID), Envelope{Data: bytes}, &resp)
	if err != nil {
		return d, err
	}

	err = json.Unmarshal(resp.Data, &d)
	if err != nil {
		return d, err
	}
	return d, err
}

func (c *Client) GetDashboard(projectName string, dashboardID string) (*Dashboard, error) {
	var (
		d    *Dashboard
		resp Envelope
	)

	err := c.CallAPI("GET", fmt.Sprintf("projects/%v/dashboards/%v", projectName, dashboardID), nil, &resp)
	if err != nil {
		return nil, err
	}

	err = json.Unmarshal(resp.Data, &d)
	if err != nil {
		return d, err
	}
	return d, err
}

func (c *Client) DeleteDashboard(projectName string, dashboardID string) error {
	err := c.CallAPI("DELETE", fmt.Sprintf("projects/%v/dashboards/%v", projectName, dashboardID), nil, nil)
	if err != nil {
		apiClientError := err.(APIResponseCarrier)
		if apiClientError.GetHTTPResponse().StatusCode != http.StatusNoContent {
			return err
		}
	}
	return nil
}
