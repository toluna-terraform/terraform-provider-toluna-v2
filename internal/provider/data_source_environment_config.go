package toluna

import (
	"context"
	"strconv"
	"time"

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
	// rules_result, _, err := kv.Get(d.Get("validation_rules").(string), nil)
	// if err != nil {
	// 	return diag.Errorf("Could not find validation rules: %s", err)
	// }
	//Get rule set to interface
	if err := d.Set("configuration", config); err != nil {
		return diag.Errorf("Could retrieve configuration: %s", err)
	}
	d.SetId(strconv.FormatInt(time.Now().Unix(), 10))
	return diags
}
