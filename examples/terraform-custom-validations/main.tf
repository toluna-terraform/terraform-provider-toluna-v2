module "validate_max_index" {
  source = "../../"
  arguments = "-a validate_min_max_env -f ${path.module}/myJsonFile.json -m 15"
}

module "validate_duplicate_index" {
  source = "../../"
  arguments = "-a validate_duplicate_index -f ${path.module}/myJsonFile.json"
}

module "validate_duplicate_env" {
  source = "../../"
  arguments = "-a validate_duplicate_env -f ${path.module}/myJsonFile.json"
}

