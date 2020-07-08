package lightstep

import (
	"fmt"
	"net/http"
)

type GetStreamAPIResponse struct {
	Data *StreamResponse `json:"data,omitempty"`
}

type ListStreamsAPIResponse struct {
	Data *ListStreamsResponse `json:"data,omitempty"`
}

type PostStreamAPIResponse struct {
	Data *StreamResponse `json:"data,omitempty"`
}

type DeleteStreamAPIResponse struct {
	Data interface{} `json:"data,omitempty"`
}

type ListStreamsResponse []StreamResponse

type StreamResponse struct {
	Response
	Attributes    StreamAttributes    `json:"attributes,omitempty"`
	Relationships StreamRelationships `json:"relationships,omitempty"`
	Links         Links               `json:"links"`
}

type StreamAttributes struct {
	Name       string                 `json:"name"`
	Query      string                 `json:"query"`
	CustomData map[string]interface{} `json:"custom-data,omitempty"`
}

type StreamRelationships struct {
	Project    LinksObj `json:"project"`
	Conditions LinksObj `json:"conditions,omitempty"`
}

type CreateOrUpdateStreamBody struct {
	Data *CreateOrUpdateStreamRequest `json:"data"`
}

type CreateOrUpdateStreamRequest struct {
	Response
	Attributes StreamRequestAttributes `json:"attributes,omitempty"`
}

type StreamRequestAttributes struct {
	Name       string                 `json:"name"`
	Query      string                 `json:"query,omitempty"`
	CustomData map[string]interface{} `json:"custom_data,omitempty"`
}

func (c *Client) CreateStream(
	projectName string,
	name string,
	query string,
	customData map[string]interface{},
) (PostStreamAPIResponse, error) {
	resp := PostStreamAPIResponse{}
	err := c.CallAPI("POST", fmt.Sprintf("projects/%v/streams", projectName), CreateOrUpdateStreamBody{
		Data: &CreateOrUpdateStreamRequest{
			Response: Response{
				Type: "stream",
			},
			Attributes: StreamRequestAttributes{
				Name:       name,
				Query:      query,
				CustomData: customData,
			},
		},
	}, &resp)

	return resp, err
}

func (c *Client) ListStreams(projectName string) (ListStreamsAPIResponse, error) {
	resp := ListStreamsAPIResponse{}

	err := c.CallAPI("GET", fmt.Sprintf("projects/%v/streams", projectName), nil, &resp)

	return resp, err
}

func (c *Client) GetStream(projectName string, StreamID string) (GetStreamAPIResponse, error) {
	resp := GetStreamAPIResponse{}
	err := c.CallAPI("GET", fmt.Sprintf("projects/%v/streams/%v", projectName, StreamID), nil, &resp)
	return resp, err
}

func (c *Client) DeleteStream(projectName string, StreamID string) error {
	err := c.CallAPI("DELETE", fmt.Sprintf("projects/%v/streams/%v", projectName, StreamID), nil, nil)
	if err != nil {
		apiClientError := err.(APIResponseCarrier)
		if apiClientError.GetHTTPResponse().StatusCode != http.StatusNoContent {
			return err
		}
	}
	return nil

}
