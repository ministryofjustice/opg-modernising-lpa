locals {
  name_prefix = "${data.aws_default_tags.current.tags.environment-name}-${data.aws_region.current.name}"
}

variable "ecs_execution_role" {
  type = object({
    id  = string
    arn = string
  })
  description = "ID and ARN of the task execution role that the Amazon ECS container agent and the Docker daemon can assume."
}

variable "ecs_task_role_arn" {
  type        = string
  description = "ARN of IAM role that allows your Amazon ECS container task to make calls to other AWS services."
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

variable "ecs_capacity_provider" {
  type        = string
  description = "Name of the capacity provider to use. Valid values are FARGATE_SPOT and FARGATE"
}

variable "ecs_application_log_group_name" {
  description = "The AWS Cloudwatch Log Group resource for application logging"
}

variable "app_service_repository_url" {
  type        = string
  description = "(optional) describe your variable"
}

variable "app_service_container_version" {
  type        = string
  description = "(optional) describe your variable"
}
