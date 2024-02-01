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

func TestSAMLGroupMappings(t *testing.T) {
	var server *httptest.Server
	server = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/public/v0.2/blars/saml-group-mappings", r.URL.Path)

		resp, err := json.Marshal(map[string]any{
			"data": map[string]any{
				"attributes": SAMLGroupMappings{
					Mappings: []SAMLGroupMapping{
						{
							SAMLAttributeKey:   "member_of",
							SAMLAttributeValue: "sre",
							OrganizationRole:   "Organization Editor",
							ProjectRoles: map[string]string{
								"project with spaces": "Project Viewer",
							},
						},
					},
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
	sgm, err := c.ListSAMLGroupMappings(context.Background())
	assert.NoError(t, err)

	assert.Equal(t, SAMLGroupMappings{
		Mappings: []SAMLGroupMapping{
			{
				SAMLAttributeKey:   "member_of",
				SAMLAttributeValue: "sre",
				OrganizationRole:   "Organization Editor",
				ProjectRoles: map[string]string{
					"project with spaces": "Project Viewer",
				},
			},
		},
	}, sgm)
}
