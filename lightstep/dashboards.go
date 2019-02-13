package lightstep

import (
	"fmt"
	"net/http"
)

type DashboardAPIResponse struct {
	Data *DashboardResponse `json:"data"`
}

type DashboardResponse struct {
	Response
	Attributes    DashboardAttributes    `json:"attributes,omitempty"`
	Relationships DashboardRelationships `json:"relationships,omitempty"`
	Links         Links                  `json:"links"`
}

type DashboardAttributes struct {
	Name     string           `json:"name"`
	Searches []SearchResponse `json:"searches"`
}

type DashboardRelationships struct {
	Project LinksObj `json:"project"`
}

type ListDashboardsAPIResponse struct {
	Data *ListDashboardsResponse `json:"data,omitempty"`
}

type ListDashboardsResponse []DashboardResponse

type DashboardRequestBody struct {
	Data *DashboardRequest `json:"data"`
}

type DashboardRequest struct {
	Response
	Attributes    DashboardRequestAttributes    `json:"attributes"`
	Relationships DashboardRequestRelationships `json:"relationships"`
}

type DashboardRequestAttributes struct {
	Name     string           `json:"name"`
	Searches []SearchResponse `json:"searches"`
}

type DashboardRequestRelationships struct {
	Dashboard ResourceIDObject `json:"dashboard"`
}

func (c *Client) CreateDashboard(
	projectName string,
	dashboardName string,
	searchAttributes []SearchAttributes,
) (DashboardAPIResponse, error) {

	resp := DashboardAPIResponse{}
	req := DashboardRequestBody{
		Data: &DashboardRequest{
			Attributes: DashboardRequestAttributes{
				Name: dashboardName,
			},
		},
	}
	for _, sa := range searchAttributes {
		req.Data.Attributes.Searches = append(
			req.Data.Attributes.Searches,
			SearchResponse{
				Attributes: sa,
			})
	}

	err := c.CallAPI("POST", fmt.Sprintf("projects/%v/dashboards", projectName), req, &resp)
	return resp, err
}

func (c *Client) GetDashboard(projectName string, dashboardID string) (DashboardAPIResponse, error) {

	resp := DashboardAPIResponse{}
	err := c.CallAPI("GET", fmt.Sprintf("projects/%v/dashboards/%v", projectName, dashboardID), nil, &resp)
	return resp, err
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
