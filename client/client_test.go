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
