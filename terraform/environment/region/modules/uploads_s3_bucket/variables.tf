variable "bucket_name" {
  description = "Name of the bucket. do not use dots (.) except for buckets that are used only for static website hosting."
}

variable "force_destroy" {
  description = "A boolean that indicates all objects should be deleted from the bucket so that the bucket can be destroyed without error. These objects are not recoverable."
  default     = false
}

variable "s3_replication_target_bucket_arn" {
  description = "The ARN of the S3 bucket where you want Amazon S3 to store replicas of the object identified by the rule."
}

variable "replication_target_encryption_key_arn" {
  description = "The Amazon Resource Name (ARN) of the AWS KMS key that Amazon S3 should use to encrypt replicas of this object."
}
