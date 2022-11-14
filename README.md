Adding custom validations to Terraform [Terraform module](https://registry.terraform.io/modules/toluna-terraform/validations/latest)

### Description
This module supports adding custom validations not supported by out of the box Terraform validations upon plan.
This is achieved by running a bash script containing custom functions , that can be call wit h different arguments,
the arguments should include the -a|--action flag which calls the function (action = function name) and any other flags required by the specific function.


## Usage

```hcl
#The following example validates there are no duplicate environments under two different data layers:
  required_providers {
    toluna = {
      source = "toluna-terraform/toluna-v2"
    } 
  }
}

data "toluna_environment_config" "app_json" {
    address = "consul-cluster-test.consul.1234546-abcd-efgh-ijkl-12345678.aws.hashicorp.cloud"
    scheme  = "https"
    path    = "terraform/app-name/app-env.json"
    validation_rules = "terraform/validations/app-config.json"
}

```

## Toggles
#### Validate arguments:

## Requirements

## Providers

| Name | Version |
|------|---------|
| <a name="toluna"></a> [assert](https://github.com/toluna-terraform/terraform-provider-toluna-v2) | >= 1.0.2 |


## Modules

## Resources

## DataSource
| Name | Source | Version |
|------|--------|---------|
| <a name="toluna_environment_config"></a> [toluna_environment_config](#data\toluna_environment_config) | ../../ |  |

No Resources.

## Inputs

No inputs.

## Outputs

No outputs.
