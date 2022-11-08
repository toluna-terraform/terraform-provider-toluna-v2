Adding custom validations to Terraform [Terraform module](https://registry.terraform.io/modules/toluna-terraform/validations/latest)

### Description
This module supports adding custom validations not supported by out of the box Terraform validations upon plan.
This is achieved by running a bash script containing custom functions , that can be call wit h different arguments,
the arguments should include the -a|--action flag which calls the function (action = function name) and any other flags required by the specific function.


## Usage

```hcl
#The following example validates there are no duplicate environments under two different data layers:
module "validate" {
  source                = "toluna-terraform/validations/custom"
  version               = "~>0.0.1" // Change to the required version.
  arguments = "-a validate_duplicate_env -f ${path.module}/some_json_file.json"
}
#The following example validates you cannot enter a negative value as an index number or an index higher then maximum possible ciders in a json file:
module "validate" {
  source                = "toluna-terraform/validations/custom"
  version               = "~>0.0.1" // Change to the required version.
  arguments = "-a validate_min_max_env -f ${path.module}/some_json_file.json -m 15"
}
#The following example validates you cannot enter a duplicate index number in a json file:
module "validate" {
  source                = "toluna-terraform/validations/custom"
  version               = "~>0.0.1" // Change to the required version.
  arguments = "-a validate_duplicate_index -f ${path.module}/some_json_file.json"
}
```

## Toggles
#### Validate arguments:
```yaml
arguments   = command line arguments to pass to the validation script I.E. -a funcation name to run -f some file to validate
```

## Requirements

| Name | Version |
|------|---------|
| <a name="requirement_terraform"></a> [terraform](#requirement\_terraform) | >= 1.0.0 |


## Providers

| Name | Version |
|------|---------|
| <a name="assert"></a> [assert](https://github.com/bwoznicki/terraform-provider-assert) | >= 0.0.1 |


## Modules

| Name | Source | Version |
|------|--------|---------|
| <a name="validate"></a> [validate](#module\validate) | ../../ |  |

## Resources

No Resources.

## Inputs

No inputs.

## Outputs

No outputs.
