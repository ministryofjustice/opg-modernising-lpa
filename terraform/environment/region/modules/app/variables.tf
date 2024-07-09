locals {
  name_prefix = "${data.aws_default_tags.current.tags.environment-name}-${data.aws_region.current.name}"
}

locals {
  policy_region_prefix = lower(replace(data.aws_region.current.name, "-", ""))
}

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
  type        = string
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

variable "app_allowed_api_arns" {
  type = list(string)
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

variable "aws_rum_guest_role_arn" {
  type        = string
  description = "ARN of the AWS RUM guest role"
  nullable    = true
}

variable "rum_monitor_application_id_secretsmanager_secret_arn" {
  type        = string
  description = "ARN of the AWS Secrets Manager secret containing the RUM monitor application ID"
  nullable    = true
}

variable "uploads_s3_bucket" {
  type = object({
    bucket_name = string
    bucket_arn  = string
  })
  description = "Name and ARN of the S3 bucket for uploads"
}

variable "event_bus" {
  type = object({
    name = string
    arn  = string
  })
  description = "Name and ARN of the event bus to send events to"
}

variable "uid_base_url" {
  type = string
}

variable "lpa_store_base_url" {
  type = string
}

variable "mock_onelogin_enabled" {
  type = bool
}

variable "mock_pay_enabled" {
  type = bool
}

variable "fault_injection_experiments_enabled" {
  type        = bool
  description = "Enable fault injection"
}

variable "search_endpoint" {
  type        = string
  description = "URL of the OpenSearch Service endpoint to use"
  nullable    = true
}

variable "search_index_name" {
  type        = string
  description = "Name of the OpenSearch Service index to use"
}

variable "search_collection_arn" {
  type        = string
  description = "ARN of the OpenSearch collection to use"
  nullable    = true
}

variable "waf_alb_association_enabled" {
  type        = bool
  description = "Enable WAF association with the ALB"
  default     = true
}
