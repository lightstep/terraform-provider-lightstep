package lightstep

import (
"fmt"
"net/http"
)

type ReadProjectsAPIResponse struct {
	Data *ProjectResponse `json:"data,omitempty"`
}

type ListProjectsAPIResponse struct {
	Data *ListProjectsResponse `json:"data,omitempty"`
}

type CreateProjectAPIResponse struct {
	Data *ProjectResponse `json:"data,omitempty"`
}

type ListProjectsResponse []ProjectResponse

type ProjectResponse struct {
	Response
	Attributes    projectAttributes    `json:"attributes,omitempty"`
	Relationships projectRelationships `json:"relationships,omitempty"`
	Links         Links                `json:"links"`
}

type projectAttributes struct {
	Name string `json:"name"`
}

type projectRelationships struct {
	Searches             LinksObj `json:"searches,omitempty"`
	NotificationPolicies LinksObj `json:"notification-policies"`
}

type CreateProjectRequest struct {
	Response
	Name string `json:"name,omitempty"`
}

type CreateProjectBody struct {
	Data *CreateProjectRequest `json:"data"`
}

type DeleteProjectAPIResponse struct {
	Data interface{} `json:"data,omitempty"`
}

func (c *Client) CreateProject(projectName string) (CreateProjectAPIResponse, error) {
	resp := CreateProjectAPIResponse{}
	err := c.CallAPI("POST", "projects", CreateProjectBody{
		Data: &CreateProjectRequest{
			Name: projectName,
		},
	}, &resp)
	return resp, err
}

func (c *Client) ReadProject(projectName string) (ReadProjectsAPIResponse, error) {
	readProjectResp := ReadProjectsAPIResponse{}
	err := c.CallAPI("GET", fmt.Sprintf("projects/%v", projectName), nil, &readProjectResp)
	return readProjectResp, err
}

func (c *Client) ListProjects() (ListProjectsAPIResponse, error) {
	resp := ListProjectsAPIResponse{}
	err := c.CallAPI("GET", "projects", nil, &resp)
	return resp, err
}

func (c *Client) DeleteProject(projectName string)error {
	resp := DeleteProjectAPIResponse{}
	err := c.CallAPI("DELETE", fmt.Sprintf("projects/%v", projectName), nil, &resp)
	if err != nil {
		apiClientError := err.(APIResponseCarrier)
		if apiClientError.GetHTTPResponse().StatusCode != http.StatusNoContent {
			return err
		}
	}
	return nil
}
