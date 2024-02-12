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

type WireEventQueryAttributes struct {
	Attributes EventQueryAttributes `json:"attributes"`
}

func (c *Client) GetEventQuery(ctx context.Context, projectName string, eventQueryID string) (*EventQueryAttributes, error) {
	var (
		event *WireEventQueryAttributes
		resp  Envelope
	)

	if err := c.CallAPI(ctx, "GET", fmt.Sprintf("projects/%v/event_queries/%v",
		projectName, eventQueryID), nil, &resp); err != nil {
		return nil, err
	}
	if err := json.Unmarshal(resp.Data, &event); err != nil {
		return nil, err
	}
	return &event.Attributes, nil
}

func (c *Client) CreateEventQuery(ctx context.Context, projectName string, attributes EventQueryAttributes) (*EventQueryAttributes, error) {
	var (
		event *EventQueryAttributes
		resp  Envelope
	)

	body := WireEventQueryAttributes{Attributes: attributes}
	bytes, err := json.Marshal(body)
	if err != nil {
		return event, err
	}
	if err := c.CallAPI(ctx, "POST",
		fmt.Sprintf("projects/%v/event_queries", projectName), Envelope{Data: bytes}, &resp); err != nil {
		return nil, err
	}
	err = json.Unmarshal(resp.Data, &event)
	return event, err
}

func (c *Client) UpdateEventQuery(ctx context.Context, projectName string, eventQueryID string, attributes EventQueryAttributes) (*EventQueryAttributes, error) {
	var (
		event *EventQueryAttributes
		resp  Envelope
	)

	bytes, err := json.Marshal(attributes)
	if err != nil {
		return event, err
	}
	if err := c.CallAPI(ctx, "PUT",
		fmt.Sprintf("projects/%v/event_queries/%v", eventQueryID, projectName), bytes, &resp); err != nil {
		return nil, err
	}
	err = json.Unmarshal(resp.Data, &event)
	return event, err
}

func (c *Client) DeleteEventQuery(ctx context.Context, projectName string, eventQueryID string) error {
	err := c.CallAPI(ctx, "DELETE",
		fmt.Sprintf("projects/%v/event_queries/%v", projectName, eventQueryID),
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
