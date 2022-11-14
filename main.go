package main

import (
	"context"
	"flag"
	"log"
	toluna "terraform-provider-toluna/internal/provider"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/plugin"
)

func main() {

	var debugMode bool
	flag.BoolVar(&debugMode, "debug", false, "set to true to run the provider with support for debuggers like delve")
	flag.Parse()
	opts := &plugin.ServeOpts{
		ProviderFunc: func() *schema.Provider {
			return toluna.New()
		},
		ProviderAddr: "registry.terraform.io/toluna-terraform/toluna-v2",
	}
	if debugMode {
		err := plugin.Debug(context.Background(), "toluna.com/edu/toluna-v2", opts)
		if err != nil {
			log.Fatal(err.Error())
		}
		return
	}
	log.SetFlags(log.Flags() &^ (log.Ldate | log.Ltime))
	plugin.Serve(opts)
}
