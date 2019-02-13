package lightstep_sdk

import "fmt"

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

func (c *PublicAPIClient) CallCreateProject(apiKey string, orgName string, projectName string) (CreateProjectAPIResponse, error) {
	resp := CreateProjectAPIResponse{}
	err := c.CallAPI(
		"POST",
		fmt.Sprintf("%v/projects", orgName),
		apiKey,
		CreateProjectBody{
			Data: &CreateProjectRequest{
				Name: projectName,
			},
		},
		&resp,
	)
	return resp, err
}

func (c *PublicAPIClient) CallReadProject(apiKey string, orgName string, projectName string) (ReadProjectsAPIResponse, error) {
	readProjectResp := ReadProjectsAPIResponse{}
	err := c.CallAPI(
		"GET",
		fmt.Sprintf("%v/projects/%v", orgName, projectName),
		apiKey,
		nil,
		&readProjectResp,
	)
	return readProjectResp, err
}

func (c *PublicAPIClient) CallListProjects(apiKey string, orgName string) (ListProjectsAPIResponse, error) {
	resp := ListProjectsAPIResponse{}
	err := c.CallAPI(
		"GET",
		fmt.Sprintf("%v/projects/", orgName),
		apiKey,
		nil,
		&resp,
	)
	return resp, err
}

func (c *PublicAPIClient) CallDeleteProject(apiKey string, orgName string, projectName string) (DeleteProjectAPIResponse, error) {
	resp := DeleteProjectAPIResponse{}
	err := c.CallAPI(
		"DELETE",
		fmt.Sprintf("%v/projects/%v", orgName, projectName),
		apiKey,
		nil,
		&resp,
	)
	return resp, err
}
