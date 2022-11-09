package toluna

import (
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

func Provider() *schema.Provider {
	return &schema.Provider{
		Schema: map[string]*schema.Schema{},
		ResourcesMap: map[string]*schema.Resource{
			"toluna_invoke_lambda":   resourceInvokeLambda(),
			"toluna_start_codebuild": resourceStartCodebuild(),
		},
		DataSourcesMap: map[string]*schema.Resource{
			"toluna_validate_configuration": dataSourceValidateConfiguration(),
			"toluna_validate_modules":       dataSourceValidateModules(),
		},
	}
}
