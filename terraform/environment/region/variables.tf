locals {
  name_prefix                   = "${data.aws_default_tags.current.tags.application}-${data.aws_default_tags.current.tags.environment-name}-${data.aws_region.current.name}"
  dns_namespace_for_environment = var.account_name == "production" ? "" : "${data.aws_default_tags.current.tags.environment-name}."
  certificate_wildcard          = var.account_name == "production" ? "" : "*."
}

variable "account_name" {
  type        = string
  description = "Name of the target account for deployments"
}
