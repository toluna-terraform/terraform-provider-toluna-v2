package main

import (
	"context"
	"encoding/json"
	"flag"
	"io/ioutil"
	"log"
	"os"
	toluna "terraform-provider-toluna/internal/provider"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/plugin"
)

type Modules struct {
	Modules []Module `json:"Modules"`
}

type Module struct {
	Key     string `json:"Key"`
	Version string `json:"Version"`
	Source  string `json:"Source"`
}

func init() {
	dir, err := os.Getwd()
	if err != nil {
		panic(err)
	}
	jsonFile, err := os.Open(dir + "/.terraform/modules/modules.json")
	if err != nil {
		panic(err)
	}
	defer jsonFile.Close()

	byteValue, _ := ioutil.ReadAll(jsonFile)
	var modules Modules
	json.Unmarshal(byteValue, &modules)
	module_versions := make([]map[string]interface{}, 0)
	for _, value := range modules.Modules {
		module := make(map[string]interface{})
		module["name"] = value.Key
		module["version"] = value.Version
		module["path"] = value.Source
		//log.Println("Test")
		//log.Fatalf("Fatal")
		module_versions = append(module_versions, module)
	}

	// diag.Errorf("Test")
	// if err := d.Set("module_versions", module_versions); err != nil {
	// 	panic(err)
	// }
	//d.SetId(strconv.FormatInt(time.Now().Unix(), 10))
	//return diags
	// dir, err := os.Getwd()
	// if err != nil {
	// 	fmt.Println(err)
	// 	return
	// }
	// d1 := `
	// data "toluna_validate_modules" "app_json" {}
	// output "name" {
	// value = data.toluna_validate_modules.app_json.module_versions
	// }`
	// f, err := os.Create(dir + "/validation.tf")
	// if err != nil {
	// 	fmt.Println(err)
	// 	return
	// }
	// f.WriteString(d1)
	// if err != nil {
	// 	fmt.Println(err)
	// 	f.Close()
	// 	return
	// }
	// err = f.Close()
	// if err != nil {
	// 	fmt.Println(err)
	// 	return
	// }

	//os.WriteFile(dir+"/validation.tf", []byte(d1), 0644)
}

func main() {
	var debugMode bool
	flag.BoolVar(&debugMode, "debug", false, "set to true to run the provider with support for debuggers like delve")
	flag.Parse()

	opts := &plugin.ServeOpts{
		ProviderFunc: func() *schema.Provider {
			return toluna.Provider()
		},
	}

	if debugMode {
		err := plugin.Debug(context.Background(), "toluna.com/edu/toluna-v2", opts)
		if err != nil {
			log.Fatal(err.Error())
		}
		return
	}

	plugin.Serve(opts)
}
