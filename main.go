package main

import (
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/plugin"
	"github.com/lightstep/terraform-provider-lightstep/exporter"
	"github.com/lightstep/terraform-provider-lightstep/lightstep"
	"log"
	"os"
)

func main() {
	if len(os.Args) > 1 && os.Args[1] == "exporter" {
		if err := exporter.Run(os.Args...); err != nil {
			log.Printf("[ERROR] %s", err.Error())
			os.Exit(1)
		}
		return
	}
	plugin.Serve(&plugin.ServeOpts{
		ProviderFunc: func() *schema.Provider {
			return lightstep.Provider()
		},
	})
}
