variable "lambda_function_image_ecr_url" {
  type = string
}

variable "lambda_function_image_tag" {
  type = string
}

variable "event_bus" {
  type = any
}

variable "schedule_runner_lambda_role" {
  type = any
}

variable "vpc_config" {
  description = "Configuration block for VPC"
  type = object({
    subnet_ids         = list(string)
    security_group_ids = list(string)
  })
}

variable "lpas_table" {
  type = object({
    arn  = string
    name = string
  })
}

variable "search_endpoint" {
  type        = string
  description = "URL of the OpenSearch Service endpoint to use"
}

variable "search_index_name" {
  type        = string
  description = "Name of the OpenSearch Service index to use"
}

variable "schedule_runner_scheduler" {
  description = "IAM role for AWS schedule runner EventBridge Scheduler"
  type        = any
}

variable "lpa_store_base_url" {
  type = string
}

variable "lpa_store_secret_arn" {
  type = string
}

variable "app_public_url" {
  type = string
}
