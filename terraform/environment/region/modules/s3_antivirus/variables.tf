variable "alarm_sns_topic_arn" {
  description = "ARN of the SNS topic for alarm notifications"
  type        = string
}

# variable "aws_subnet_ids" {
#   type        = list(string)
#   description = "List of Sirius private subnet Ids"
# }

variable "data_store_bucket" {
  type        = any
  description = "Data store bucket to scan for viruses"
}

variable "definition_bucket" {
  type        = any
  description = "Bucket containing virus definitions"
}

variable "ecr_image_uri" {
  type        = string
  description = "URI of ECR image to use for Lambda"
}

variable "environment_variables" {
  description = "A map that defines environment variables for the Lambda Function."
  type        = map(string)
  default     = {}
}

variable "lambda_task_role" {
  type        = any
  description = "Execution role for Lambda"
}

variable "s3_antivirus_provisioned_concurrency" {
  description = "Number of concurrent executions to provision for Lambda"
  type        = number
}
