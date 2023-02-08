package client

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
)

func Test_DeleteDestionation_when_connection_is_closed(t *testing.T) {
	var server *httptest.Server
	server = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/public/v0.2/blars/projects/tacoman/destinations/hi", r.URL.Path)

		server.CloseClientConnections()
	}))
	defer server.Close()

	t.Setenv("LIGHTSTEP_API_BASE_URL", server.URL)
	c := NewClient("api", "blars", "staging")
	err := c.DeleteDestination(context.Background(), "tacoman", "hi")

	assert.NotNil(t, err)

	apiErr, ok := err.(APIClientError)

	assert.True(t, ok)
	assert.Equal(t, -1, apiErr.GetStatusCode())
}

func Test_DeleteDestination_when_connection_has_wrong_content_length(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Length", "1") // set content length to 1 so reading body fails
	}))
	defer server.Close()

	t.Setenv("LIGHTSTEP_API_BASE_URL", server.URL)
	c := NewClient("api", "blars", "staging")
	err := c.DeleteDestination(context.Background(), "tacoman", "hi")

	assert.NotNil(t, err)
	assert.Equal(t, "unexpected EOF", err.Error())
}
