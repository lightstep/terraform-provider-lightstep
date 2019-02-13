package lightstep_sdk

import (
	"fmt"
	"net/http"
)

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

func (c *Client) CreateSearch(
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

func (c *Client) ListSearches(
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

func (c *Client) GetSearch(apiKey string, organizationName string, projectName string, searchID string) (GetSearchAPIResponse, error) {
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

func (c *Client) DeleteSearch(apiKey string, organizationName string, projectName string, searchID string) error {
	err := c.CallAPI(
		"DELETE",
		fmt.Sprintf("%v/projects/%v/searches/%v", organizationName, projectName, searchID),
		apiKey,
		nil,
		nil,
	)
	if err != nil {
		apiClientError := err.(APIResponseCarrier)
		if apiClientError.GetHTTPResponse().StatusCode != http.StatusNoContent {
			return err
		}
	}
	return nil

}
