package lightstep

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

var testAccProviders map[string]*schema.Provider
var testAccProvider *schema.Provider
var testAccProviderFactories map[string]func() (*schema.Provider, error)

func init() {

	testAccProvider = Provider()
	testAccProviders = map[string]*schema.Provider{
		"lightstep": testAccProvider,
	}

	testAccProviderFactories = map[string]func() (*schema.Provider, error){
		"lightstep": func() (*schema.Provider, error) { return Provider(), nil }, //nolint:unparam
	}
}

func TestProvider_impl(t *testing.T) {
	var _ *schema.Provider = Provider()
}

func TestProvider(t *testing.T) {
	if err := Provider().InternalValidate(); err != nil {
		t.Fatalf("err: %s", err)
	}
}
