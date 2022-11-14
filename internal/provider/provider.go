package toluna

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"strings"
	toluna "terraform-provider-toluna/utils"

	"github.com/hashicorp/go-version"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

type getItemsRequest struct {
	SortBy     string
	SortOrder  string
	ItemsToGet int
}

type getItemsResponseError struct {
	Message string `json:"message"`
}

type getItemsResponseData struct {
	Item string `json:"item"`
}

type getItemsResponseBody struct {
	Result string                 `json:"result"`
	Data   []getItemsResponseData `json:"data"`
	Error  getItemsResponseError  `json:"error"`
}

type getItemsResponseHeaders struct {
	ContentType string `json:"Content-Type"`
}

type getItemsResponse struct {
	StatusCode int                     `json:"statusCode"`
	Headers    getItemsResponseHeaders `json:"headers"`
	Body       getItemsResponseBody    `json:"body"`
}

var (
	config struct {
		Paths []string
	}
)

func init() {
	schema.DescriptionKind = schema.StringMarkdown

	schema.SchemaDescriptionBuilder = func(s *schema.Schema) string {
		descriptionWithDefault := s.Description
		if s.Default != nil {
			descriptionWithDefault += fmt.Sprintf(" Defaults to `%v`.", s.Default)
		}
		return strings.TrimSpace(descriptionWithDefault)
	}
}

func New() *schema.Provider {

	return &schema.Provider{
		Schema: map[string]*schema.Schema{
			"strict_module_validation": &schema.Schema{
				Type:     schema.TypeBool,
				Optional: true,
				Default:  false,
			},
		},
		ResourcesMap: map[string]*schema.Resource{
			"toluna_invoke_lambda":   resourceInvokeLambda(),
			"toluna_start_codebuild": resourceStartCodebuild(),
		},
		DataSourcesMap: map[string]*schema.Resource{
			"toluna_environment_config": dataSourceEnvironmentConfig(),
		},
		ConfigureContextFunc: providerConfigure,
	}
}
func providerConfigure(ctx context.Context, d *schema.ResourceData) (interface{}, diag.Diagnostics) {
	var diags diag.Diagnostics
	var severity diag.Severity
	var message string
	strict_module_validation := d.Get("strict_module_validation").(bool)
	if strict_module_validation {
		severity = 0
		message = "You must updated to latest version to continue working."
	} else {
		severity = 1
		message = "It is advised to update to latest version."
	}
	dir, err := os.Getwd()
	if err != nil {
		panic(err)
	}
	config.Paths = append(config.Paths, dir)
	out, err := toluna.ScanModules(config.Paths)
	if err != nil {
		fmt.Println("failed to retrieve modules")
	}
	for _, r := range out {
		if strings.HasPrefix(r.ModuleCall.Source, "toluna-terraform") {
			remoteVersion := GetRemoteVersion(r.ModuleCall.Source)
			localVersion := r.ModuleCall.Version
			// Remove no-semver chars
			localVersion = strings.ReplaceAll(localVersion, ">", "")
			localVersion = strings.ReplaceAll(localVersion, "~", "")
			localVersion = strings.ReplaceAll(localVersion, "=", "")
			result := CompareVersions(remoteVersion, localVersion)
			if result == true {
				diags = append(diags, diag.Diagnostic{
					Severity: severity,
					Summary:  "Module outdated: " + r.ModuleCall.Name + ", Under: " + r.Path,
					Detail:   "A newer module version of " + r.ModuleCall.Name + " [ " + localVersion + " ]" + " Is available [ " + remoteVersion + " ], " + message,
				})
			}
		}
	}
	return nil, diags
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
	return lv.LessThan(rv)
}
