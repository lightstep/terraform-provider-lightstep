package client

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func Test_UserRoleBinding(t *testing.T) {
	var server *httptest.Server
	server = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/public/v0.2/blars/role-binding", r.URL.Path)

		assert.Equal(t, "project with spaces", r.URL.Query().Get("project"))
		assert.Equal(t, "project editor", r.URL.Query().Get("role-name"))

		resp, err := json.Marshal(map[string]any{
			"data": map[string]any{
				"attributes": RoleBinding{
					RoleName:    "project editor",
					ProjectName: "project with spaces",
					Users:       []string{"user1@lightstep.com"},
				},
			},
		})
		require.NoError(t, err)

		w.Write(resp)
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	t.Setenv("LIGHTSTEP_API_BASE_URL", server.URL)
	c := NewClient("api", "blars", "staging")
	rb, err := c.ListRoleBinding(context.Background(), "project with spaces", "project editor")
	assert.NoError(t, err)

	assert.Equal(t, RoleBinding{
		RoleName:    "project editor",
		ProjectName: "project with spaces",
		Users:       []string{"user1@lightstep.com"},
	}, rb)
}
