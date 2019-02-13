package lightstep_sdk

import "fmt"

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
	apiKey string,
	orgName string,
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

	err := c.CallAPI(
		"POST",
		fmt.Sprintf("%v/projects/%v/dashboards", orgName, projectName),
		apiKey,
		req,
		&resp,
	)
	return resp, err
}
