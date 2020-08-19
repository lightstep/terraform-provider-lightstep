package lightstep

import (
	"encoding/json"
	"fmt"
	"net/http"
)

type DestinationType string

const (
	PAGERDUTY_DESTINATION_TYPE DestinationType = "pagerduty"
	WEBHOOK_DESTINATION_TYPE   DestinationType = "webhook"
	SLACK_DESTINATION_TYPE     DestinationType = "slack"
)

type Destination struct {
	Type       string      `json:"type"`
	Attributes interface{} `json:"attributes"`
	ID         string      `json:"id"`
}

type WebhookAttributes struct {
	Name            string                 `json:"name"`
	DestinationType DestinationType        `json:"destination_type"`
	URL             string                 `json:"url"`
	CustomHeaders   map[string]interface{} `json:"custom_headers,omitempty"`
}

func (c *Client) CreateDestination(
	project string,
	destination Destination) (Destination, error) {

	var resp Envelope
	var dest Destination

	bytes, err := json.Marshal(destination)
	if err != nil {
		return dest, err
	}

	err = c.CallAPI("POST", fmt.Sprintf("projects/%v/destinations", project), Envelope{Data: bytes}, &resp)
	if err != nil {
		return dest, err
	}

	err = json.Unmarshal(resp.Data, &dest)
	return dest, err
}

func (c *Client) GetDestination(projectName string, destinationID string) (Destination, error) {
	var (
		dest Destination
		resp Envelope
	)

	err := c.CallAPI("GET",
		fmt.Sprintf("projects/%v/destinations/%v", projectName, destinationID),
		nil,
		&resp)

	if err != nil {
		return dest, err
	}

	err = json.Unmarshal(resp.Data, &dest)

	return dest, err
}

func (c *Client) DeleteDestination(project string, destinationID string) error {
	err := c.CallAPI("DELETE", fmt.Sprintf("projects/%v/destinations/%v", project, destinationID), nil, nil)
	if err != nil {
		apiClientError := err.(APIResponseCarrier)
		if apiClientError.GetHTTPResponse().StatusCode != http.StatusNoContent {
			return err
		}
	}
	return nil
}
