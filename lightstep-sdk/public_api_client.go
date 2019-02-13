package lightstep_sdk

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"net/http/cookiejar"
	"time"
)

// ***************************************************************
// Structs generally useful for Public APIs
// ***************************************************************

type Response struct {
	Type string `json:"type"`
	ID   string `json:"id"`
}

type ResourceIDObject struct {
	Response
}

// Serializes as `{ "links": { "key" : "value" } }`
type LinksObj struct {
	Links Links `json:"links,omitempty"`
}

type RelatedLinkObj struct {
	HREF string `json:"href,omitempty"`
	Text string `json:"text,omitempty"`
}

type Links map[string]interface{}

type Body struct {
	Data   interface{}            `json:"data,omitempty"`
	Errors []string               `json:"errors,omitempty"`
	Links  map[string]interface{} `json:"links,omitempty"`
}

type PublicAPIClient struct {
	hostname    string
	client      *http.Client
	contentType string
}

// ***************************************************************
// Structs for Condition APIs
// ***************************************************************
type ConditionAPIResponse struct {
	Data *ConditionResponse
}

type ConditionResponse struct {
	Response
	Attributes    ConditionAttributes    `json:"attributes,omitempty"`
	Relationships ConditionRelationships `json:"relationships,omitempty"`
	Links         Links                  `json:"links"`
}

type ConditionAttributes struct {
	Name               *string                `json:"name"`
	EvaluationWindowMs int64                  `json:"eval-window-ms"`
	Expression         string                 `json:"expression"`
	CustomData         map[string]interface{} `json:"custom-data,omitempty"`
}

type ConditionRelationships struct {
	Project LinksObj `json:"project"`
	Search  LinksObj `json:"search"`
}

type ListConditionsAPIResponse struct {
	Data *ListConditionsResponse `json:"data,omitempty"`
}

type ListConditionsResponse []ConditionResponse

type ConditionStatusAPIResponse struct {
	Data *ConditionStatusResponse
}

type ConditionStatusResponse struct {
	Response
	Attributes    ConditionStatusAttributes    `json:"attributes,omitempty"`
	Relationships ConditionStatusRelationships `json:"relationships,omitempty"`
}

type ConditionStatusAttributes struct {
	Expression  string `json:"expression"`
	State       string `json:"state"`
	Description string `json:"description"`
}

type ConditionStatusRelationships struct {
	Condition LinksObj `json:"condition"`
}

type ConditionRequestBody struct {
	Data *ConditionRequest `json:"data"`
}

type ConditionRequest struct {
	Response
	Attributes    ConditionRequestAttributes    `json:"attributes"`
	Relationships ConditionRequestRelationships `json:"relationships"`
}

type ConditionRequestAttributes struct {
	Name               *string                 `json:"name"`
	Expression         *string                 `json:"expression"`
	EvaluationWindowMs *int64                  `json:"eval-window-ms"`
	CustomData         *map[string]interface{} `json:"custom-data"`
}

type ConditionRequestRelationships struct {
	Search ResourceIDObject `json:"search"`
}

// ***************************************************************
// Structs for Project APIs
// ***************************************************************
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

// ***************************************************************
// Structs for ServiceDirectory APIs
// ***************************************************************
type ListServicesAPIResponse struct {
	Data *ListServicesResponse `json:"data,omitempty"`
}

type ListServicesResponse struct {
	Type  string            `json:"type"`
	Links Links             `json:"links"`
	Items []ServiceResponse `json:"items"`
}

type ListServicesRequest struct {
	Offset *int `json:"offset"`
	Limit  *int `json:"limit"`
}

type ServiceResponse struct {
	ID            string               `json:"id"`
	Attributes    ServiceAttributes    `json:"attributes,omitempty"`
	Relationships ServiceRelationships `json:"relationships,omitempty"`
}

type ServiceAttributes struct {
	Name    string    `json:"name"`
	Created time.Time `json:"created"`
	Updated time.Time `json:"updated"`
}

type ServiceRelationships struct {
	Project LinksObj `json:"project"`
}

// ***************************************************************
// Structs for Search APIs
// ***************************************************************

type GetSearchAPIResponse struct {
	Data *SearchResponse `json:"data,omitempty"`
}

type ListSearchesAPIResponse struct {
	Data *ListSearchesResponse `json:"data,omitempty"`
}

type PostSearchAPIResponse struct {
	Data *SearchResponse `json:"data,omitempty"`
}

type DeleteSearchAPIResponse struct {
	Data interface{} `json:"data,omitempty"`
}

type ListSearchesResponse []SearchResponse

type SearchResponse struct {
	Response
	Attributes    SearchAttributes    `json:"attributes,omitempty"`
	Relationships SearchRelationships `json:"relationships,omitempty"`
	Links         Links               `json:"links"`
}

type SearchAttributes struct {
	Name       string                 `json:"name"`
	Query      string                 `json:"query"`
	CustomData map[string]interface{} `json:"custom-data,omitempty"`
}

type SearchRelationships struct {
	Project    LinksObj `json:"project"`
	Conditions LinksObj `json:"conditions,omitempty"`
}

type CreateOrUpdateSearchBody struct {
	Data *CreateOrUpdateSearchRequest `json:"data"`
}

type CreateOrUpdateSearchRequest struct {
	Response
	Attributes SearchRequestAttributes `json:"attributes,omitempty"`
}

type SearchRequestAttributes struct {
	Name       string                 `json:"name"`
	Query      string                 `json:"query,omitempty"`
	CustomData map[string]interface{} `json:"custom_data,omitempty"`
}

// ***************************************************************
// Structs for Users API
// ***************************************************************
type CreateUserAPIResponse struct {
	Data *CreateUserResponse `json:"data"`
}

type CreateUserResponse struct {
	Response
	LinksObj
	Attributes CreateUserResponseAttributes `json:"attributes,omitempty"`
}

type CreateUserResponseAttributes struct {
	Username string `json:"username"`
	Role     string `json:"role"`
}

type CreateUserRequestAttributes struct {
	Username      string `json:"username"`
	Role          string `json:"role"`
	WithLoginLink bool   `json:"with-login-link"`
}

type CreateUserRequest struct {
	Attributes CreateUserRequestAttributes `json:"attributes"`
}

type CreateUserRequestBody struct {
	Data *CreateUserRequest `json:"data"`
}

// ***************************************************************
// Structs for Internal API
// ***************************************************************
type CreateTracingAccountAPIResponse struct {
	Data *CreateTracingAccountResponse `json:"data"`
}

type CreateTracingAccountResponse struct {
	Response
	LinksObj
	Attributes CreateTracingAccountResponseAttributes `json:"attributes,omitempty"`
}

type CreateTracingAccountResponseAttributes struct {
	OrganizationName string `json:"organization_name"`
	LoginLink        string `json:"login_link"`
}

type CreateTracingAccountRequestBody struct {
	Data *CreateTracingAccountRequest `json:"data"`
}

type CreateTracingAccountRequest struct {
	Attributes CreateTracingAccountRequestAttributes `json:"attributes"`
}

type CreateTracingAccountRequestAttributes struct {
	CreatorUsername  string `json:"creator_username"`
	OrganizationName string `json:"organization_name"`
}

// ***************************************************************
// Structs for Dashboard APIs
// ***************************************************************
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

// NewPublicAPIClient gets a client for the public API
func NewPublicAPIClient(ctx context.Context, hostname string) *PublicAPIClient {
	cookieJar, err := cookiejar.New(nil)
	if err != nil {
		log.Fatal(ctx, err)
	}
	return &PublicAPIClient{
		hostname: hostname,
		client: &http.Client{
			Jar: cookieJar,
		},
		contentType: "application/vnd.api+json",
	}
}

// CallAPI calls the given API and unmarshals the result to into result.
func (c *PublicAPIClient) CallAPI(
	httpMethod string,
	suffix string,
	authToken string,
	data interface{},
	result interface{},
) error {
	return callAPI(
		context.TODO(),
		c.client,
		generateURLForPublicAPI(c.hostname, suffix),
		httpMethod,
		Headers{
			"Authorization": fmt.Sprintf("bearer %v", authToken),
			"Content-Type":  c.contentType,
			"Accept":        c.contentType,
		},
		data,
		result,
	)
}

func generateURLForPublicAPI(hostname string, suffix string) string {
	return fmt.Sprintf("%v/public/v0.1/%v", hostname, suffix)
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

func (c *PublicAPIClient) CallCreateSearch(
	apiKey string,
	orgName string,
	projectName string,
	name string,
	query string,
	customData map[string]interface{},
) (PostSearchAPIResponse, error) {
	resp := PostSearchAPIResponse{}

	err := c.CallAPI(
		"POST",
		fmt.Sprintf("%v/projects/%v/searches", orgName, projectName),
		apiKey,
		CreateOrUpdateSearchBody{
			Data: &CreateOrUpdateSearchRequest{
				Response: Response{
					Type: "search",
				},
				Attributes: SearchRequestAttributes{
					Name:  name,
					Query: query,
				},
			},
		},
		&resp)

	return resp, err
}

func (c *PublicAPIClient) CallListSearches(
	apiKey string,
	orgName string,
	projectName string,
) (ListSearchesAPIResponse, error) {
	resp := ListSearchesAPIResponse{}

	err := c.CallAPI(
		"GET",
		fmt.Sprintf("%v/projects/%v/searches", orgName, projectName),
		apiKey,
		nil,
		&resp)

	return resp, err
}

func (c *PublicAPIClient) CallGetSearch(apiKey string, organizationName string, projectName string, searchID string) (GetSearchAPIResponse, error) {
	resp := GetSearchAPIResponse{}
	err := c.CallAPI(
		"GET",
		fmt.Sprintf("%v/projects/%v/searches/%v", organizationName, projectName, searchID),
		apiKey,
		nil,
		&resp,
	)
	return resp, err
}

func (c *PublicAPIClient) CallDeleteSearch(apiKey string, organizationName string, projectName string, searchID string) (DeleteSearchAPIResponse, error) {
	resp := DeleteSearchAPIResponse{}
	err := c.CallAPI(
		"DELETE",
		fmt.Sprintf("%v/projects/%v/searches/%v", organizationName, projectName, searchID),
		apiKey,
		nil,
		&resp,
	)
	return resp, err
}

// ========= Public API helpers
func (c *PublicAPIClient) CallCreateDashboard(
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
