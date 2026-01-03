terraform {
  source = "tfr:///${local.module}?version=${local.module_version}"
}

locals {
  module         = "terraform-aws-modules/vpc/aws"
  module_version = "5.8.1"
}
