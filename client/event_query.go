package client

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
)

type EventQueryAttributes struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	QueryString string `json:"query_string"`
	Source      string `json:"source"`
	Type        string `json:"type"`
}

func (c *Client) GetEventQuery(ctx context.Context, projectName string, eventQueryID string) (*EventQueryAttributes, error) {
	var (
		dest *EventQueryAttributes
		resp Envelope
	)

	err := c.CallAPI(ctx, "GET",
		fmt.Sprintf("/projects/%v/event_queries/%v", projectName, eventQueryID),
		nil,
		&resp)

	if err != nil {
		return nil, err
	}
	err = json.Unmarshal(resp.Data, &dest)

	return dest, err
}

func (c *Client) CreateEventQuery(ctx context.Context, projectName string, attributes EventQueryAttributes) (*EventQueryAttributes, error) {
	var (
		dest *EventQueryAttributes
		resp Envelope
	)

	bytes, err := json.Marshal(attributes)
	if err != nil {
		return dest, err
	}
	if err := c.CallAPI(ctx, "POST",
		fmt.Sprintf("/projects/%v/event_queries", projectName), bytes, &resp); err != nil {
		return nil, err
	}
	err = json.Unmarshal(resp.Data, &dest)
	return dest, err
}

func (c *Client) UpdateEventQuery(ctx context.Context, projectName string, eventQueryID string, attributes EventQueryAttributes) (*EventQueryAttributes, error) {
	var (
		dest *EventQueryAttributes
		resp Envelope
	)

	bytes, err := json.Marshal(attributes)
	if err != nil {
		return dest, err
	}
	if err := c.CallAPI(ctx, "PUT",
		fmt.Sprintf("/projects/%v/event_queries/%v", eventQueryID, projectName), bytes, &resp); err != nil {
		return nil, err
	}
	err = json.Unmarshal(resp.Data, &dest)
	return dest, err
}

func (c *Client) DeleteEventQuery(ctx context.Context, projectName string, eventQueryID string) error {
	err := c.CallAPI(ctx, "DELETE",
		fmt.Sprintf("/projects/%v/event_queries/%v", projectName, eventQueryID),
		nil,
		nil)
	if err != nil {
		apiClientError, ok := err.(APIResponseCarrier)
		if !ok || apiClientError.GetStatusCode() != http.StatusNoContent {
			return err
		}
	}
	return nil
}
