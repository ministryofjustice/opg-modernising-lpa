variable "network_cidr_block" {
  type        = string
  description = "The IPv4 CIDR block for the VPC. CIDR can be explicitly set or it can be derived from IPAM using ipv4_netmask_length."
}

variable "flow_log_cloudwatch_log_group_kms_key_id" {
  type        = string
  default     = null
  description = "The ARN of the KMS Key to use when encrypting the Default VPC flow log data."
}
