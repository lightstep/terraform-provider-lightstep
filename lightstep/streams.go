package lightstep

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
)

type Stream struct {
	Type          string              `json:"type,omitempty"`
	ID            string              `json:"id,omitempty"`
	Attributes    StreamAttributes    `json:"attributes,omitempty"`
	Relationships StreamRelationships `json:"relationships,omitempty"`
	Links         Links               `json:"links,omitempty"`
}

type StreamAttributes struct {
	Name       string                 `json:"name"`
	Query      string                 `json:"query"`
	CustomData map[string]interface{} `json:"custom-data,omitempty"`
}

type StreamRelationships struct {
	Project    LinksObj `json:"project,omitempty"`
	Conditions LinksObj `json:"conditions,omitempty"`
}

func (c *Client) CreateStream(
	projectName string,
	name string,
	query string,
	customData map[string]interface{},
) (Stream, error) {
	var (
		s    Stream
		resp Envelope
	)


	bytes, err := json.Marshal(
		Stream{
		Type: "stream",
		Attributes: StreamAttributes{
			Name:       name,
			Query:      query,
			CustomData: customData,
		},
	})

	if err != nil {
		log.Printf("error marshalling data: %v", err)
	}

	req := Envelope{
		Data: bytes,
	}

	err = c.CallAPI("POST", fmt.Sprintf("projects/%v/streams", projectName), req, &resp)

	err = json.Unmarshal(resp.Data, &s)
	return s, err
}

func (c *Client) ListStreams(projectName string) ([]Stream, error) {
	var (
		s    []Stream
		resp Envelope
	)

	err := c.CallAPI("GET", fmt.Sprintf("projects/%v/streams", projectName), nil, &resp)
	if err != nil {
		log.Printf("Error: %v", err)
		return s, err
	}
	err = json.Unmarshal(resp.Data, &s)
	if err != nil {
		log.Printf("Error unmarshaling response: %v", err)
		return s, err
	}
	return s, err
}

func (c *Client) GetStream(projectName string, StreamID string) (Stream, error) {
	var (
		s Stream
	)
	resp := Envelope{}

	err := c.CallAPI("GET", fmt.Sprintf("projects/%v/streams/%v", projectName, StreamID), nil, &resp)
	if err != nil {
		log.Printf("Error: %v", err)
		return s, err
	}

	err = json.Unmarshal(resp.Data, &s)
	if err != nil {
		log.Printf("Error unmarshaling response: %v", err)
		return s, err
	}
	return s, err
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
