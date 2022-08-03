locals {
  name_prefix                   = "${data.aws_default_tags.current.tags.environment-name}-${data.aws_region.current.name}"
  dns_namespace_for_environment = var.account_name == "production" ? "" : "${data.aws_default_tags.current.tags.environment-name}."
  certificate_wildcard          = var.account_name == "production" ? "" : "*."
}

variable "account_name" {
  type        = string
  description = "Name of the target account for deployments"
}

variable "ecs_execution_role" {
  type = object({
    id  = string
    arn = string
  })
  description = "ID and ARN of the task execution role that the Amazon ECS container agent and the Docker daemon can assume."
}

variable "ecs_cluster" {
  type        = string
  description = "ARN of an ECS cluster."
}

variable "ecs_service_desired_count" {
  type        = number
  default     = 0
  description = "Number of instances of the task definition to place and keep running. Defaults to 0. Do not specify if using the DAEMON scheduling strategy."
}

variable "network" {
  type = object({
    vpc_id              = string
    application_subnets = list(string)
    public_subnets      = list(string)
  })
  description = "VPC ID, a list of application subnets, and a list of private subnets required to provision the ECS service"
}
