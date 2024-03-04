package client

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func Test_getUnifiedDashboardURL(t *testing.T) {
	testCases := []struct {
		projID   [2]string
		expected string
	}{
		{
			[2]string{"my_project", "123"},
			"projects/my_project/metric_dashboards/123",
		},
		{
			[2]string{"ProductionEnvironment", "fLx72349023"},
			"projects/ProductionEnvironment/metric_dashboards/fLx72349023",
		},
	}
	for _, c := range testCases {
		result := getUnifiedDashboardURL(c.projID[0], c.projID[1])
		require.Equal(t, c.expected, result)
	}
}

func Test_DeleteMetricDashboard_when_connection_is_closed(t *testing.T) {
	var server *httptest.Server
	server = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/public/v0.2/blars/projects/tacoman/metric_dashboards/hi", r.URL.Path)

		server.CloseClientConnections()
	}))
	defer server.Close()

	c := NewClient("api", "blars", server.URL)
	err := c.DeleteUnifiedDashboard(context.Background(), "tacoman", "hi")
	assert.NotNil(t, err)

	apiErr, ok := err.(APIClientError)

	assert.True(t, ok)
	assert.Equal(t, -1, apiErr.GetStatusCode())
}

func Test_DeleteMetricDashboard_when_connection_has_wrong_content_length(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Length", "1") // set content length to 1 so reading body fails
	}))
	defer server.Close()

	c := NewClient("api", "blars", server.URL)
	err := c.DeleteUnifiedDashboard(context.Background(), "tacoman", "hi")

	assert.NotNil(t, err)
	assert.Equal(t, "unexpected EOF", err.Error())
}
