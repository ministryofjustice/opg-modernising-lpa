variable "network_cidr_block" {
  type        = string
  description = "The IPv4 CIDR block for the VPC. CIDR can be explicitly set or it can be derived from IPAM using ipv4_netmask_length."
}

variable "cloudwatch_log_group_kms_key_alias" {
  type        = string
  default     = null
  description = "The alias of the KMS Key to use when encrypting Cloudwatch log data."
}

variable "sns_kms_key" {
  type        = any
  description = "The KMS key used to encrypt the SNS topic"
}

variable "secrets_manager_kms_key_alias" {
  description = "The alias of the KMS key used to encrypt Secrets Manager secrets"
  type        = string
}

variable "reduced_fees_uploads_s3_encryption_kms_key_alias" {
  description = "The alias of the KMS key used to encrypt the reduced fees uploads S3 bucket and replication manifests"
  type        = string
}

variable "dynamodb_exports_s3_bucket_server_side_encryption_key_id" {
  description = "The ID of the KMS key to use for server-side encryption of the bucket."
  type        = string
}

variable "opensearch_kms_target_key_arn" {
  type        = string
  description = "kms target key ARN for opensearch serverless encryption"
}

variable "pagerduty_service_name" {
  type = string
}

variable "athena_enabled" {
  type        = bool
  description = "Enable athena for this account"
}

variable "athena_s3_target_key_id" {
  description = "The ID of the KMS key to use for server-side encryption of the bucket."
  type        = string
}
