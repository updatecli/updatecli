terraform {}

locals {}

generate "main" {
  path      = "main.tf"
  if_exists = "overwrite_terragrunt"
  contents  = <<EOF
variable "tags" {
  type = map(string)
  default = {}
}
EOF
}