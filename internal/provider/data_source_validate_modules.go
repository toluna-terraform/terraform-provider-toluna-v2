package toluna

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/hashicorp/go-version"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

type Modules struct {
	Modules []Module `json:"Modules"`
}

type Module struct {
	Key     string `json:"Key"`
	Version string `json:"Version"`
	Source  string `json:"Source"`
}

func dataSourceValidateModules() *schema.Resource {
	return &schema.Resource{
		ReadContext: dataSourceValidateModulesRead,
		Schema: map[string]*schema.Schema{
			"module_versions": &schema.Schema{
				Type:     schema.TypeList,
				Computed: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"name": &schema.Schema{
							Type:     schema.TypeString,
							Computed: true,
						},
						"version": &schema.Schema{
							Type:     schema.TypeString,
							Computed: true,
						},
						"path": &schema.Schema{
							Type:     schema.TypeString,
							Computed: true,
						},
					},
				},
			},
		},
	}
}

func dataSourceValidateModulesRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
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
	for _, value := range modules.Modules {
		if strings.HasPrefix(value.Source, "registry.terraform.io/toluna-terraform") {
			officialModuleName := strings.TrimPrefix(value.Source, "registry.terraform.io/")
			latestVersion := GetRemoteVersion(officialModuleName)
			result := CompareVersions(latestVersion, value.Version)
			if result == false {
				diags = append(diags, diag.Diagnostic{
					Severity: diag.Error,
					Summary:  "Module outdated",
					Detail:   "A newer module version of " + value.Key + "[" + value.Version + "]" + " Is available [" + latestVersion + "], You must updated to latest version to continue working.",
				})
			}
		}
	}

	d.SetId(strconv.FormatInt(time.Now().Unix(), 10))
	return diags
}

func GetRemoteVersion(moduleName string) string {
	baseURL := "https://registry.terraform.io/v1/modules/"
	client := &http.Client{}
	req, _ := http.NewRequest("GET", baseURL+moduleName, nil)
	req.Header.Add("Accept", "application/json")
	resp, err := client.Do(req)
	if err != nil {
		fmt.Println("Error reading terraform registry")
	}

	defer resp.Body.Close()
	resp_body, _ := ioutil.ReadAll(resp.Body)
	module_version := make(map[string]interface{})
	if err := json.Unmarshal(resp_body, &module_version); err != nil { // Parse []byte to the go struct pointer
		fmt.Println("Can not get remote version")
	}
	return module_version["version"].(string)
}

func CompareVersions(remoteVersion, localVersion string) bool {
	lv, _ := version.NewVersion(localVersion)
	rv, _ := version.NewVersion(remoteVersion)
	if lv.LessThan(rv) {
		return false
	}
	return true
}
