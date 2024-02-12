package client

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNew_public(t *testing.T) {
	t.Parallel()
	c := NewClient("api-key", "org-name", "https://api.lightstep.com")
	assert.Equal(t, "https://api.lightstep.com/public/v0.2/org-name", c.baseUrl)
}

func TestNew_env_var_provided_baseURL(t *testing.T) {
	// Parallel not used here due to t.Setenv.
	t.Setenv("LIGHTSTEP_API_BASE_URL", "http://localhost:8080")
	c := NewClient("api-key", "org-name", "")
	assert.Equal(t, "http://localhost:8080/public/v0.2/org-name", c.baseUrl)
}
