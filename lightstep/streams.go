package lightstep

import (
	"encoding/json"
	"fmt"
	"net/http"
)

type Stream struct {
	Type       string           `json:"type,omitempty"`
	ID         string           `json:"id,omitempty"`
	Attributes StreamAttributes `json:"attributes,omitempty"`
}

type StreamAttributes struct {
	Name       string                 `json:"name"`
	Query      string                 `json:"query"`
	CustomData map[string]interface{} `json:"custom-data,omitempty"`
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
		return s, err
	}

	err = c.CallAPI(
		"POST",
		fmt.Sprintf("projects/%v/streams", projectName),
		Envelope{Data: bytes},
		&resp)
	if err != nil {
		return s, err
	}

	err = json.Unmarshal(resp.Data, &s)
	if err != nil {
		return s, err
	}

	return s, err
}

func (c *Client) ListStreams(projectName string) ([]Stream, error) {
	var (
		s    []Stream
		resp Envelope
	)

	err := c.CallAPI("GET", fmt.Sprintf("projects/%v/streams", projectName), nil, &resp)
	if err != nil {
		return s, err
	}
	err = json.Unmarshal(resp.Data, &s)
	if err != nil {
		return s, err
	}
	return s, err
}

func (c *Client) GetStream(projectName string, StreamID string) (Stream, error) {
	var (
		s    Stream
		resp Envelope
	)

	err := c.CallAPI("GET", fmt.Sprintf("projects/%v/streams/%v", projectName, StreamID), nil, &resp)
	if err != nil {
		return s, err
	}

	err = json.Unmarshal(resp.Data, &s)
	if err != nil {
		return s, err
	}
	return s, err
}

func (c *Client) UpdateStream(projectName string,
	streamID string,
	stream Stream,
) (Stream, error) {

	var (
		s    Stream
		resp Envelope
	)

	bytes, err := json.Marshal(&stream)
	if err != nil {
		return s, err
	}

	err = c.CallAPI("PATCH", fmt.Sprintf("projects/%v/streams/%v", projectName, streamID), Envelope{Data: bytes}, &resp)
	if err != nil {
		return s, err
	}

	err = json.Unmarshal(resp.Data, &s)
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
