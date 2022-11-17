package client

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
)

type (
	Notebook struct {
		ID         string
		Attributes NotebookAttributes     `json:"attributes,omitempty"`
		Links      map[string]interface{} `json:"links"`
	}

	NotebookAttributes struct {
		Name        string          `json:"name"`
		Description string          `json:"description"`
		Entries     []NotebookEntry `json:"entries"`
	}

	NotebookEntry struct {
		ID string `json:"id"`
		// The type of entry. Currently the two available types are "chart" and "text_block".
		Type        string `json:"type"`
		StartMicros int64  `json:"start_micros"`
		EndMicros   int64  `json:"end_micros"`
		Rank        int    `json:"rank"`

		// Only one of the following fields is set. The field that is set depends on the Type.
		TextBlock *NotebookTextBlock `json:"text_block"`
		Chart     *NotebookChart     `json:"chart"`
	}

	NotebookTextBlock struct {
		Text string `json:"text"`
	}

	NotebookChart struct {
		ID           string   `json:"id"`
		Title        string   `json:"title"`
		ChartType    string   `json:"chart_type"`
		YAxis        *YAxis   `json:"y_axis"`
		QueryStrings []string `json:"query-strings"`
		Subtitle     *string  `json:"subtitle"`
	}
)

func (c *Client) CreateNotebook(
	ctx context.Context,
	projectID string,
	notebook Notebook,
) (Notebook, error) {

	var resp Envelope
	var respNotebook Notebook

	bytes, err := json.Marshal(Notebook{
		Attributes: notebook.Attributes,
	})
	if err != nil {
		return respNotebook, err
	}

	err = c.CallAPI(ctx, "POST", getNotebookURL(projectID, ""), Envelope{Data: bytes}, &resp)
	if err != nil {
		return respNotebook, err
	}

	err = json.Unmarshal(resp.Data, &respNotebook)
	return respNotebook, err
}

func (c *Client) GetNotebook(ctx context.Context, projectID string, notebookID string) (*Notebook, error) {
	var (
		respNotebook *Notebook
		resp         Envelope
	)

	err := c.CallAPI(ctx,
		"GET",
		getNotebookURL(projectID, notebookID),
		nil,
		&resp,
	)

	if err != nil {
		return nil, err
	}
	err = json.Unmarshal(resp.Data, &respNotebook)

	return respNotebook, err
}

func (c *Client) UpdateNotebook(
	ctx context.Context,
	projectID string,
	notebookID string,
	notebook Notebook,
) (Notebook, error) {

	var resp Envelope
	var respNotebook Notebook

	bytes, err := json.Marshal(Notebook{
		ID:         notebookID,
		Attributes: notebook.Attributes,
	})
	if err != nil {
		return respNotebook, err
	}

	err = c.CallAPI(ctx, "PUT", getNotebookURL(projectID, notebook.ID), Envelope{Data: bytes}, &resp)
	if err != nil {
		return respNotebook, err
	}

	err = json.Unmarshal(resp.Data, &respNotebook)
	return respNotebook, err
}

func (c *Client) DeleteNotebook(ctx context.Context, projectID string, notebookID string) error {
	err := c.CallAPI(ctx, "DELETE", getNotebookURL(projectID, notebookID), nil, nil)
	if err != nil {
		apiClientError := err.(APIResponseCarrier)
		if apiClientError.GetHTTPResponse().StatusCode != http.StatusNoContent {
			return err
		}
	}
	return nil
}

func getNotebookURL(projectID string, notebookID string) string {
	base := fmt.Sprintf("projects/%s/notebooks", projectID)

	if notebookID != "" {
		return fmt.Sprintf("%s/%s", base, notebookID)
	}
	return base
}
