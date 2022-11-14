
terraform {
  required_providers {
    toluna = {
      source = "toluna-terraform/toluna"
      version = ">=0.0.9"
    }
  }
}

provider "aws" {
  region  = "us-east-1"
  profile = "my-profile"
}

provider "toluna" {
  strict_module_validation = true
}


resource "toluna_invoke_lambda" "example" {
  region = "us-east-1"
  aws_profile = "my-profile"
  function_name = "my_lambda"
  payload = jsonencode({"name": "example pay load"})
}

resource "toluna_start_codebuild" "example" {
  region = "us-east-1"
  aws_profile = "my-profile"
  project_name = "my_project"
    environment_variables  {
    name = "my-variable"
    value = "FOO"
    type = "PLAINTEXT"
  }
  environment_variables  {
    name = "my-secret-variable"
    value = "BAR"
    type = "PARAMETER_STORE"
  }
  environment_variables  {
    name = "my-other-secret-variable"
    value = "BAR"
    type = "SECRETS_MANAGER"
  }
}

data "toluna_environment_config" "app_json" {
    address = "consul-cluster-test.consul.1234546-abcd-efgh-ijkl-12345678.aws.hashicorp.cloud"
    scheme  = "https"
    path    = "terraform/app-name/app-env.json"
    validation_rules = "terraform/validations/app-config.json"
}

locals {
  env_vars         = jsondecode("${data.toluna_environment_config.app_json.configuration}")[local.env_name]
}