package exporter

import (
	"bytes"
	"testing"

	"github.com/lightstep/terraform-provider-lightstep/client"
)

func TestExportToHCL(t *testing.T) {

	export := func(d *client.UnifiedDashboard) (string, error) {
		var buf bytes.Buffer
		err := exportToHCL(&buf, d)
		return string(buf.Bytes()), err
	}

	t.Run("Test QueryString Export", func(t *testing.T) {
		export(nil)
	})
}
