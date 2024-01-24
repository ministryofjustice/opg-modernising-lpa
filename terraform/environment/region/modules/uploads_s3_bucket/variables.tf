variable "bucket_name" {
  description = "Name of the bucket. do not use dots (.) except for buckets that are used only for static website hosting."
  type        = string
}

variable "force_destroy" {
  description = "A boolean that indicates all objects should be deleted from the bucket so that the bucket can be destroyed without error. These objects are not recoverable."
  default     = false
  type        = bool
}

variable "s3_replication" {
  type = object({
    enabled                                   = bool
    destination_bucket_arn                    = string
    destination_encryption_key_arn            = string
    destination_account_id                    = string
    lambda_function_image_ecr_arn             = string
    lambda_function_image_ecr_url             = string
    lambda_function_image_tag                 = string
    enable_s3_batch_job_replication_scheduler = bool
  })
  description = <<EOT
    s3_replication = {
      enabled                                   = "Enable S3 object replication"
      destination_bucket_arn                    = "ARN of the destination bucket"
      destination_encryption_key_arn            = "ARN of the destination encryption key"
      destination_account_id                    = "Account ID of the destination bucket"
      lambda_function_image_ecr_arn             = "ARN of the lambda function to be invoked on a schedule to create replication jobs"
      lambda_function_image_ecr_url             = "URL of the lambda function to be invoked on a schedule to create replication jobs"
      lambda_function_image_tag                 = "Tag of the lambda function to be invoked on a schedule to create replication jobs"
      enable_s3_batch_job_replication_scheduler = "Enable scheduler to create replication jobs"
    }
    EOT
}

variable "events_received_lambda_function" {
  type        = any
  description = "Lambda function ARN for events received"
}

variable "s3_antivirus_lambda_function" {
  type        = any
  description = "Lambda function ARN for events received"
}
