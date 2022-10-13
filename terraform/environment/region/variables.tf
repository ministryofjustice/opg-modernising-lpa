locals {
  name_prefix = "${data.aws_default_tags.current.tags.environment-name}-${data.aws_region.current.name}"
}

variable "ecs_execution_role" {
  type        = any
  description = "The task execution role that the Amazon ECS container agent and the Docker daemon can assume."
}

variable "ecs_task_roles" {
  type = object({
    app = any
  })
  description = "ARN of IAM role that allows your Amazon ECS container task to make calls to other AWS services."
}

variable "application_log_retention_days" {
  type        = number
  description = "Specifies the number of days you want to retain log events in the specified log group. Possible values are: 1, 3, 5, 7, 14, 30, 60, 90, 120, 150, 180, 365, 400, 545, 731, 1827, 3653, and 0. If you select 0, the events in the log group are always retained and never expire."
}

variable "ecs_capacity_provider" {
  type        = string
  description = "Name of the capacity provider to use. Valid values are FARGATE_SPOT and FARGATE"
}

variable "app_service_repository_url" {
  type        = string
  description = "(optional) describe your variable"
}

variable "app_service_container_version" {
  type        = string
  description = "(optional) describe your variable"
}

variable "ingress_allow_list_cidr" {
  type        = list(string)
  description = "List of CIDR ranges permitted to access the service"
}

variable "alb_deletion_protection_enabled" {
  type        = bool
  description = "If true, deletion of the load balancer will be disabled via the AWS API. This will prevent Terraform from deleting the load balancer. Defaults to false."
}

variable "lpas_table" {
  type        = any
  description = "DynamoDB table for storing LPAs"
}

variable "app_env_vars" {
  type        = any
  description = "Environment variable values for app"
}

variable "public_access_enabled" {
  type        = bool
  description = "Enable access to the Modernising LPA service from the public internet"
}
