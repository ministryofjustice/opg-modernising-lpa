variable "lambda_function_image_ecr_url" {
  type = string
}

variable "lambda_function_image_tag" {
  type = string
}

variable "event_bus_name" {
  type = string
}

variable "allowed_api_arns" {
  type = list(string)
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

variable "search_collection_arn" {
  type        = string
  description = "ARN of the OpenSearch collection to use"
  nullable    = true
}

variable "uploads_bucket" {
  type = any
}

variable "event_bus_arn" {
  type = string
}
