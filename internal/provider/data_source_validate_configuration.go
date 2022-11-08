package toluna

import (
	"context"
	"encoding/json"
	"fmt"
	"regexp"
	"strconv"
	"strings"

	"github.com/PaesslerAG/jsonpath"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func dataSourceValidateConfiguration() *schema.Resource {
	return &schema.Resource{
		ReadContext: dataSourceValidateConfigurationRead,
		Schema: map[string]*schema.Schema{
			"json_config": {
				Type:     schema.TypeMap,
				Required: true,
			},
			"rule_set": {
				Type:     schema.TypeSet,
				Required: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"key_name": {
							Type:     schema.TypeString,
							Required: true,
						},
						"rule": {
							Type:     schema.TypeString,
							Required: true,
							ValidateFunc: func(val any, key string) (warns []string, errs []error) {
								v := val.(string)
								rules := map[string]bool{
									"==":       true,
									"!=":       true,
									">":        true,
									"<":        true,
									"<=":       true,
									">=":       true,
									"contains": true,
									"odd":      true,
									"even":     true,
									"unique":   true,
								}
								if !rules[v] {
									errs = append(errs, fmt.Errorf("%q must be one of the following rules [== | != | > | < | => | <= | contains | odd | even | unique] got: %d", key, v))
								}
								return
							},
						},
						"value": {
							Type:     schema.TypeString,
							Required: true,
						},
					},
				},
			},
		},
	}
}
func Even(number int) bool {
	return number%2 == 0
}

func Odd(number int) bool {
	return !Even(number)
}
func getValue(config map[string]interface{}, elem string, key_name string) []string {
	if key_name == "key" {
		v := make(map[string]interface{})
		json.Unmarshal([]byte(elem), &v)
		k := make([]string, len(config))
		for s, _ := range v {
			k = append(k, s)
		}
		return k
	} else {
		v := interface{}(nil)
		json.Unmarshal([]byte(elem), &v)
		values, err := jsonpath.Get(key_name, v)
		if err != nil {
			fmt.Println(err)
			return nil
		}
		var k []string
		for _, value := range values.([]interface{}) {
			val := strings.ReplaceAll(fmt.Sprintln(value), "\n", "")
			k = append(k, val)
		}
		return k
	}
}

func dataSourceValidateConfigurationRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {

	// // Warning or errors can be collected in a slice type
	var diags diag.Diagnostics
	//'add validation and return result
	configuration := d.Get("json_config").(map[string]interface{})
	validation_rule := d.Get("rule_set").(*schema.Set)
	for key, element := range configuration {
		fmt.Println("%b", key)
		elem := make(map[string]interface{})
		json.Unmarshal([]byte(element.(string)), &elem)
		rules := validation_rule.List()
		for i := range rules {
			rules_map := rules[i].(map[string]interface{})
			key_name := rules_map["key_name"].(string)
			rule := rules_map["rule"].(string)
			value := rules_map["value"].(string)
			if key_name == "" || rule == "" || value == "" {
				return diag.Errorf("missing param: %s", rules[i])
			}
			value_arr := getValue(configuration, element.(string), key_name)
			if len(value_arr) == 0 {
				return diag.Errorf("key: %s not found", key_name)
			}
			for v := range value_arr {
				switch {
				case rule == "==":
					if value_arr[v] != value {
						return diag.Errorf("Key %s in %s is not equal to:  %s", key_name, value_arr, value)
					}
				case rule == "!=":
					if value_arr[v] == value {
						return diag.Errorf("Key %s in %s should not be equal to:  %s", key_name, value_arr, value)
					}
				case rule == ">":
					intVar, err := strconv.Atoi(value_arr[v])
					intValue, err := strconv.Atoi(value)
					if err != nil {
						return diag.Errorf("Cannot compare string with number %s", value_arr[v])
					}
					if intVar <= intValue {
						return diag.Errorf("Key %s in %s should be greater then: %s", key_name, value_arr, value)
					}
				case rule == "<":
					intVar, err := strconv.Atoi(value_arr[v])
					intValue, err := strconv.Atoi(value)
					if err != nil {
						return diag.Errorf("Cannot compare string with number %s", value_arr[v])
					}
					if intVar >= intValue {
						return diag.Errorf("Key %s in %s should be lower then: %s", key_name, value_arr, value)
					}
				case rule == ">=":
					intVar, err := strconv.Atoi(value_arr[v])
					intValue, err := strconv.Atoi(value)
					if err != nil {
						return diag.Errorf("Cannot compare string with number %s", value_arr[v])
					}
					if intVar < intValue {
						return diag.Errorf("Key %s in %s should be greater then or equal to: %s", key_name, value_arr, value)
					}
				case rule == "<=":
					intVar, err := strconv.Atoi(value_arr[v])
					intValue, err := strconv.Atoi(value)
					if err != nil {
						return diag.Errorf("Cannot compare string with number %s", value_arr[v])
					}
					if intVar > intValue {
						return diag.Errorf("Key %s in %s should be lower then or equal to: %s", key_name, value_arr, value)
					}
				case rule == "contains":
					for _, s := range value_arr {
						if strings.Contains(s, value) {
							return diag.Errorf("Key %s in %s should not contain:  %s", key_name, value_arr, value)
						}
					}
				case rule == "odd":
					intVar, err := strconv.Atoi(value_arr[v])
					if err != nil {
						return diag.Errorf("Cannot compare string with number %s", value_arr[v])
					}
					if Even(intVar) {
						return diag.Errorf("Key %s with value %s is an even number but should be odd", key_name, value_arr[v])
					}
				case rule == "even":
					intVar, err := strconv.Atoi(value_arr[v])
					if err != nil {
						return diag.Errorf("Cannot compare string with number %s", value_arr[v])
					}
					if Odd(intVar) {
						return diag.Errorf("Key %s with value %s is an odd number but should be even", key_name, value_arr[v])
					}
				case rule == "unique":
					if key_name == "key" {
						count := 0
						key_regexp := fmt.Sprintf(`"%v".*:`, value)
						var re = regexp.MustCompile(key_regexp)
						for i, match := range re.FindAllString(element.(string), -1) {
							fmt.Println(match, "found at index", i)
							count++
						}
						if count > 1 {
							return diag.Errorf("Value %s in %s for Key %s is used more then once", value, value_arr, key_name)
						}
					}
					occurred := map[string]bool{}
					for v := range value_arr {
						if occurred[value_arr[v]] != true {
							occurred[value_arr[v]] = true
						} else {
							return diag.Errorf("Value %s in %s for Key %s is used more then once", value, value_arr, key_name)
						}
					}
				}
			}
		}
	}
	return diags
}
