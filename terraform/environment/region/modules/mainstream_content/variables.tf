variable "ecs_execution_role" {
  type = object({
    id  = string
    arn = string
  })
  description = "ID and ARN of the task execution role that the Amazon ECS container agent and the Docker daemon can assume."
}

variable "ecs_task_role" {
  type        = any
  description = "ARN of IAM role that allows your Amazon ECS container task to make calls to other AWS services."
}

variable "ecs_cluster" {
  type        = string
  description = "ARN of an ECS cluster."
}

variable "ecs_service_desired_count" {
  type        = number
  default     = 1
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
  type        = string
  description = "The AWS Cloudwatch Log Group resource for application logging"
}

variable "mrlpa_content_repository_url" {
  type        = string
  description = "(optional) describe your variable"
}

variable "mrlpa_content_container_sha_digest" {
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

variable "container_port" {
  type        = number
  description = "Port on the container to associate with."
}

variable "public_access_enabled" {
  type        = bool
  description = "Enable access to the Modernising LPA service from the public internet"
}

variable "ecs_cpu_architecture" {
  type        = string
  description = "The CPU architecture to use for the task. The valid values are 'X86_64' and 'ARM64'. Defaults to 'X86_64'."
  validation {
    condition     = contains(["X86_64", "ARM64"], var.ecs_cpu_architecture)
    error_message = "Invalid CPU architecture. Valid values are 'X86_64' and 'ARM64'."
  }
}

variable "mrlpa_service_url" {
  type        = string
  description = "MRLPA URL"
}
