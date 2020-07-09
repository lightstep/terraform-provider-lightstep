package lightstep

import (
	"encoding/json"
	"fmt"
	"net/http"
)

type Dashboard struct {
	Type          string                 `json:"type,omitempty"`
	ID            string                 `json:"id,omitempty"`
	Attributes    DashboardAttributes    `json:"attributes,omitempty"`
	Relationships DashboardRelationships `json:"relationships,omitempty"`
	Links         Links                  `json:"links"`
}

type DashboardAttributes struct {
	Name    string   `json:"name"`
	Streams []Stream `json:"streams"`
}

type DashboardRelationships struct {
	Project LinksObj `json:"project"`
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

	req := Envelope{Data: bytes}

	err = c.CallAPI("POST", fmt.Sprintf("projects/%v/dashboards", projectName), req, &resp)
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

	req := Envelope{Data: bytes}

	err = c.CallAPI("PATCH", fmt.Sprintf("projects/%v/dashboards/%v", projectName, dashboardID), req, &resp)
	if err != nil {
		return d, err
	}

	err = json.Unmarshal(resp.Data, &d)
	if err != nil {
		return d, err
	}
	return d, err
}

func (c *Client) GetDashboard(projectName string, dashboardID string) (Dashboard, error) {
	var d Dashboard
	resp := Envelope{}
	err := c.CallAPI("GET", fmt.Sprintf("projects/%v/dashboards/%v", projectName, dashboardID), nil, &resp)
	if err != nil {
		return d, err
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
