package lightstep

import (
	"fmt"
	"net/http"
)

type Stream struct {
	Type          string              `default:"stream"`
	ID            string              `json:"id,omitempty"`
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

type Links map[string]string

type LinksObj struct {
	Links Links `json:"links"`
}

type StreamAPIResponse struct {
	Data *Stream `json:"data,omitempty"`
}

type ListStreamsAPIResponse struct {
	Data []Stream `json:"data,omitempty"`
}

type CreateOrUpdateStreamBody struct {
	Data *CreateOrUpdateStreamRequest `json:"data"`
}

type CreateOrUpdateStreamRequest struct {
	Type       string           `json:"type"`
	ID         string           `json:"id,omitempty"`
	Attributes StreamAttributes `json:"attributes,omitempty"`
}

func (c *Client) CreateStream(
	projectName string,
	name string,
	query string,
	customData map[string]interface{},
) (StreamAPIResponse, error) {
	resp := StreamAPIResponse{}
	err := c.CallAPI("POST", fmt.Sprintf("projects/%v/streams", projectName), CreateOrUpdateStreamBody{
		Data: &CreateOrUpdateStreamRequest{
			Type: "stream",
			Attributes: StreamAttributes{
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

func (c *Client) GetStream(projectName string, StreamID string) (StreamAPIResponse, error) {
	resp := StreamAPIResponse{}
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
