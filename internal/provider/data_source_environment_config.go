package toluna

import (
	"context"
	"encoding/json"
	"fmt"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/PaesslerAG/jsonpath"
	"github.com/hashicorp/consul/api"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func dataSourceEnvironmentConfig() *schema.Resource {
	return &schema.Resource{
		ReadContext: dataSourceEnvironmentConfigRead,
		Schema: map[string]*schema.Schema{
			"id": &schema.Schema{
				Type:     schema.TypeInt,
				Computed: true,
			},
			"configuration": &schema.Schema{
				Type:     schema.TypeString,
				Computed: true,
			},
			"address": {
				Type:     schema.TypeString,
				Required: true,
			},
			"scheme": {
				Type:     schema.TypeString,
				Required: true,
			},
			"path": {
				Type:     schema.TypeString,
				Required: true,
			},
			"validation_rules": {
				Type:     schema.TypeString,
				Required: true,
			},
		},
	}
}

func dataSourceEnvironmentConfigRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	client, err := api.NewClient(&api.Config{
		Address: d.Get("address").(string),
		Scheme:  d.Get("scheme").(string),
	})
	if err != nil {
		return diag.Errorf("Could not connect to Consul: %s", err)
	}
	kv := client.KV()

	config_result, _, err := kv.Get(d.Get("path").(string), nil)
	if err != nil {
		return diag.Errorf("Could not find configuration: %s", err)
	}
	config := string(config_result.Value)
	rules_result, _, err := kv.Get(d.Get("validation_rules").(string), nil)
	if err != nil {
		return diag.Errorf("Could not find rule set: %s", err)
	}
	rules_list := string(rules_result.Value)

	var rules map[string]interface{}
	var configuration map[string]interface{}
	json.Unmarshal([]byte(rules_list), &rules)
	json.Unmarshal([]byte(config), &configuration)
	for rule_key, rule_value := range rules {
		fmt.Printf("%v", rule_key)
		rules_map := rule_value.(map[string]interface{})
		key_name := rules_map["key_name"].(string)
		rule := rules_map["rule"].(string)
		value := rules_map["value"].(string)
		if key_name == "" || rule == "" || value == "" {
			return diag.Errorf("missing param: %s", rule_value)
		}

		value_arr := getValue(configuration, key_name)

		if len(value_arr) == 0 {
			diags = append(diags, diag.Diagnostic{
				Severity: diag.Error,
				Summary:  "Key not found",
				Detail:   fmt.Sprintf("key: %s not found , %s", key_name, configuration),
			})
		}
		for v := range value_arr {
			switch {
			case rule == "==":
				if value_arr[v] != value {
					diags = append(diags, diag.Diagnostic{
						Severity: diag.Error,
						Summary:  "Configuration key mismatch",
						Detail:   fmt.Sprintf("Key %s in %s is not equal to:  %s", key_name, value_arr, value),
					})
				}
			case rule == "!=":
				if value_arr[v] == value {
					diags = append(diags, diag.Diagnostic{
						Severity: diag.Error,
						Summary:  "Configuration key mismatch",
						Detail:   fmt.Sprintf("Key %s in %s should not be equal to:  %s", key_name, value_arr, value),
					})
				}
			case rule == ">":
				intVar, err := strconv.Atoi(value_arr[v])
				intValue, err := strconv.Atoi(value)
				if err != nil {
					diags = append(diags, diag.Diagnostic{
						Severity: diag.Error,
						Summary:  "Invalid Configuration format",
						Detail:   fmt.Sprintf("Cannot compare string with number %s", value_arr[v]),
					})
				}
				if intVar <= intValue {
					diags = append(diags, diag.Diagnostic{
						Severity: diag.Error,
						Summary:  "Configuration key mismatch",
						Detail:   fmt.Sprintf("Key %s in %s should be greater then: %s", key_name, value_arr, value),
					})
				}
			case rule == "<":
				intVar, err := strconv.Atoi(value_arr[v])
				intValue, err := strconv.Atoi(value)
				if err != nil {
					diags = append(diags, diag.Diagnostic{
						Severity: diag.Error,
						Summary:  "Invalid Configuration format",
						Detail:   fmt.Sprintf("Cannot compare string with number %s", value_arr[v]),
					})
				}
				if intVar >= intValue {
					diags = append(diags, diag.Diagnostic{
						Severity: diag.Error,
						Summary:  "Configuration key mismatch",
						Detail:   fmt.Sprintf("Key %s in %s should be lower then: %s", key_name, value_arr, value),
					})
				}
			case rule == ">=":
				intVar, err := strconv.Atoi(value_arr[v])
				intValue, err := strconv.Atoi(value)
				if err != nil {
					diags = append(diags, diag.Diagnostic{
						Severity: diag.Error,
						Summary:  "Invalid Configuration format",
						Detail:   fmt.Sprintf("Cannot compare string with number %s", value_arr[v]),
					})
				}
				if intVar < intValue {
					diags = append(diags, diag.Diagnostic{
						Severity: diag.Error,
						Summary:  "Configuration key mismatch",
						Detail:   fmt.Sprintf("Key %s in %s should be greater then or equal to: %s", key_name, value_arr, value),
					})
				}
			case rule == "<=":
				intVar, err := strconv.Atoi(value_arr[v])
				intValue, err := strconv.Atoi(value)
				if err != nil {
					diags = append(diags, diag.Diagnostic{
						Severity: diag.Error,
						Summary:  "Invalid Configuration format",
						Detail:   fmt.Sprintf("Cannot compare string with number %s", value_arr[v]),
					})
				}
				if intVar > intValue {
					diags = append(diags, diag.Diagnostic{
						Severity: diag.Error,
						Summary:  "Configuration key mismatch",
						Detail:   fmt.Sprintf("Key %s in %s should be lower then or equal to: %s", key_name, value_arr, value),
					})
				}
			case rule == "not_contain":
				if strings.Contains(value_arr[v], value) {
					diags = append(diags, diag.Diagnostic{
						Severity: diag.Error,
						Summary:  "Configuration key mismatch",
						Detail:   fmt.Sprintf("Key %s in %s should not contain:  %s", value_arr[v], value_arr, value),
					})
				}
			case rule == "odd":
				intVar, err := strconv.Atoi(value_arr[v])
				if err != nil {
					diags = append(diags, diag.Diagnostic{
						Severity: diag.Error,
						Summary:  "Invalid Configuration format",
						Detail:   fmt.Sprintf("Cannot compare string with number %s", value_arr[v]),
					})
				}
				if Even(intVar) {
					diags = append(diags, diag.Diagnostic{
						Severity: diag.Error,
						Summary:  "Configuration key mismatch",
						Detail:   fmt.Sprintf("Key %s with value %s is an even number but should be odd", key_name, value_arr[v]),
					})
				}
			case rule == "even":
				intVar, err := strconv.Atoi(value_arr[v])
				if err != nil {
					diags = append(diags, diag.Diagnostic{
						Severity: diag.Error,
						Summary:  "Invalid Configuration format",
						Detail:   fmt.Sprintf("Cannot compare string with number %s", value_arr[v]),
					})
				}
				if Odd(intVar) {
					diags = append(diags, diag.Diagnostic{
						Severity: diag.Error,
						Summary:  "Configuration key mismatch",
						Detail:   fmt.Sprintf("Key %s with value %s is an odd number but should be even", key_name, value_arr[v]),
					})
				}
			case rule == "unique":
				if key_name == "key" {
					count := 0
					key_regexp := fmt.Sprintf(`"%v".*:`, value_arr[v])
					var re = regexp.MustCompile(key_regexp)
					for i, match := range re.FindAllString(config, -1) {
						fmt.Println(match, "found at index", i)
						count++
					}
					if count > 1 {
						diags = append(diags, diag.Diagnostic{
							Severity: diag.Error,
							Summary:  fmt.Sprintf("Key name %s is used more then once.", value_arr[v]),
							Detail:   fmt.Sprintf("%s", config),
						})
					}
				}
				if checkDup(value_arr, value_arr[v]) {
					diags = append(diags, diag.Diagnostic{
						Severity: diag.Error,
						Summary:  fmt.Sprintf("Key %s has duplicate value %s.", key_name, value_arr[v]),
						Detail:   fmt.Sprintf("%s", config),
					})
				}
			}
		}
	}

	if diags != nil {
		return diags
	}
	if err := d.Set("configuration", config); err != nil {
		return diag.Errorf("Could retrieve configuration: %s", err)
	}
	d.SetId(strconv.FormatInt(time.Now().Unix(), 10))
	return diags
}

func Even(number int) bool {
	return number%2 == 0
}

func Odd(number int) bool {
	return !Even(number)
}

func checkDup(heystack []string, needle string) bool {
	count := 0
	var i int
	for i = 0; i < len(heystack); i++ {
		if heystack[i] == needle {
			count++
		}
	}
	if count == 1 {
		return false
	} else {
		return true
	}

}

func getValue(config map[string]interface{}, key_name string) []string {
	if key_name == "key" {
		var s []string
		for k, _ := range config {
			if k != "" {
				s = append(s, k)
			}
		}
		return s
	} else {
		values, err := jsonpath.Get(key_name, config)
		if err != nil {
			fmt.Println(err)
			return nil
		}
		var k []string
		for _, value := range values.([]interface{}) {
			val := strings.ReplaceAll(fmt.Sprintln(value), "\n", "")
			if val != "" {
				k = append(k, val)
			}
		}
		return k
	}
}
