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
    enabled                        = bool
    destination_bucket_arn         = string
    destination_encryption_key_arn = string
    destination_account_id         = string
  })
  description = <<EOT
    s3_replication = {
      enabled                        = "Enable S3 object replication"
      destination_bucket_arn         = "ARN of the destination bucket"
      destination_encryption_key_arn = "ARN of the destination encryption key"
      destination_account_id         = "Account ID of the destination bucket"
    }
    EOT
}
