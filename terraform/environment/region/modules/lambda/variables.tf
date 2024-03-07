variable "environment" {
  description = "The environment lambda is being deployed to."
  type        = string
}

variable "memory" {
  description = "The memory to use."
  type        = number
  default     = null
}

variable "image_uri" {
  description = "The image uri in ECR."
  type        = string
  default     = null
}

variable "description" {
  description = "Description of your Lambda Function (or Layer)"
  type        = string
  default     = null
}

variable "environment_variables" {
  description = "A map that defines environment variables for the Lambda Function."
  type        = map(string)
  default     = {}
}

variable "lambda_name" {
  description = "A unique name for your Lambda Function"
  type        = string
}

variable "package_type" {
  description = "The Lambda deployment package type."
  type        = string
  default     = "Image"
}

variable "timeout" {
  description = "The amount of time your Lambda Function has to run in seconds."
  type        = number
  default     = 30
}

variable "kms_key" {
  type        = any
  description = "KMS key for the lambda log group"
}

variable "iam_policy_documents" {
  description = "List of IAM policy documents that are merged together. Documents later in the list override earlier ones"
  type        = list(string)
  default     = []
}

variable "aws_iam_role" {
  description = "The IAM role for the lambda"
  type        = any
}

variable "vpc_config" {
  description = "Configuration block for VPC"
  type = object({
    subnet_ids         = list(string)
    security_group_ids = list(string)
  })
  default  = null
  nullable = true
}
