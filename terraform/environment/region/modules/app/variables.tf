locals {
  name_prefix                   = "${data.aws_default_tags.current.tags.environment-name}-${data.aws_region.current.name}"
  dns_namespace_for_environment = var.account_name == "production" ? "" : "${data.aws_default_tags.current.tags.environment-name}."
  certificate_wildcard          = var.account_name == "production" ? "" : "*."
}

variable "account_name" {
  type        = string
  description = "Name of the target account for deployments"
}

variable "ecs_execution_role_arn" {
  type        = string
  description = "ARN of the task execution role that the Amazon ECS container agent and the Docker daemon can assume."
}
