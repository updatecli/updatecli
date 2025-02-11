terraform {
  source = "tfr://${local.module}?version=1.2.3"
}

locals {
  module = "terraform-aws-modules/auroravpc/aws"
}
