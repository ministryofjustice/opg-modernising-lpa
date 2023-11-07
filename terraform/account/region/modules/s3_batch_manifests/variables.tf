variable "s3_encryption_kms_key_alias" {
  description = "The alias of the KMS key used to encrypt the reduced fees uploads S3 bucket and replication manifests"
  type        = string
}
