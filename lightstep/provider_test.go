package lightstep

import (
	"os"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

var testAccProviders map[string]*schema.Provider
var testAccProvider *schema.Provider
var testAccProviderFactories map[string]func() (*schema.Provider, error)
var testProject string

func init() {
	testAccProvider = Provider()
	testAccProviders = map[string]*schema.Provider{
		"lightstep": testAccProvider,
	}

	testAccProviderFactories = map[string]func() (*schema.Provider, error){
		"lightstep": func() (*schema.Provider, error) { return Provider(), nil }, //nolint:unparam
	}

	testProject = os.Getenv("LIGHTSTEP_PROJECT")
	if testProject == "" {
		testProject = "terraform-provider-tests"
	}
}

func TestProvider(t *testing.T) {
	if err := Provider().InternalValidate(); err != nil {
		t.Fatalf("err: %s", err)
	}
}
