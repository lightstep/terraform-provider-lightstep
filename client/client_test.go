package client

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNew_public(t *testing.T) {
	t.Parallel()

	c := NewClient("api-key", "org-name", "public")
	assert.Equal(t, "https://api.lightstep.com/public/v0.2/org-name", c.baseURL)
}

func TestNew_other(t *testing.T) {
	t.Parallel()
	c := NewClient("api-key", "org-name", "other")
	assert.Equal(t, "https://api-other.lightstep.com/public/v0.2/org-name", c.baseURL)
}

func TestNew_env_var_provided_baseURL(t *testing.T) {
	// Parallel not used here due to t.Setenv.
	t.Setenv("LIGHTSTEP_API_BASE_URL", "http://localhost:8080")
	c := NewClient("api-key", "org-name", "public")
	assert.Equal(t, "http://localhost:8080/public/v0.2/org-name", c.baseURL)
}

func Test_resolveBaseURL(t *testing.T) {
	t.Parallel()
	assert.Equal(t, "https://api.lightstep.com", resolveBaseURL("public"))
	assert.Equal(t, "https://api-meta.lightstep.com", resolveBaseURL("meta"))
	assert.Equal(t, "https://api.eu.lightstep.com", resolveBaseURL("eu-public"))
	assert.Equal(t, "https://api.eu-staging.lightstep.com", resolveBaseURL("eu-staging"))
	assert.Equal(t, "https://api-other.lightstep.com", resolveBaseURL("other"))
}
