package lightstep

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"

	"github.com/lightstep/terraform-provider-lightstep/client"
)

func TestAccStreamDatasource(t *testing.T) {
	streamConfig := `
resource "lightstep_stream" "aggie_errors_ds" {
  project_name = ` + fmt.Sprintf("\"%s\"", testProject) + `
  stream_name = "Aggie Errors DS"
  query = "service IN (\"aggie_ds\") AND \"error\" IN (\"true\")"
}

data "lightstep_stream" "stream_ds" {
  	depends_on = [
    	lightstep_stream.aggie_errors_ds,
  	]
	project_name = ` + fmt.Sprintf("\"%s\"", testProject) + `
	stream_id = lightstep_stream.aggie_errors_ds.id
}
`
	var stream client.Stream
	resource.Test(t, resource.TestCase{
		PreCheck:  func() { testAccPreCheck(t) },
		Providers: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: streamConfig,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckStreamExists("lightstep_stream.aggie_errors_ds", &stream),
					resource.TestCheckResourceAttr("data.lightstep_stream.stream_ds", "stream_name", "Aggie Errors DS"),
					resource.TestCheckResourceAttr("data.lightstep_stream.stream_ds", "stream_query", "service IN (\"aggie_ds\") AND \"error\" IN (\"true\")"),
				),
			},
		},
	})
}
