terraform {
  source = local.base_source_url
}

locals {
  base_source_url = "tfr://terraform-aws-modules/aurora/aws?version=5.8.1"
}
