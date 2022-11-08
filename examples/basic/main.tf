
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

data "toluna_validate_configuration" "example" {
  rule_set {
    key_name = "key"
    rule ="unique"
    value = "nil" 
  }
  rule_set {
    key_name = "$..env_index"
    rule ="odd"
    value = "nil" 
  }
  rule_set {
    key_name = "$..env_index"
    rule =">"
    value = "6" 
  }
  rule_set {
    key_name = "$..env_index"
    rule ="<"
    value = "21" 
  }
  rule_set {
    key_name = "key"
    rule ="~="
    value = "example" 
  }
  json_config = data.consul_keys.appjson.var
} 