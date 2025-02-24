variable "target_event_bus_arns" {
  type        = map(string)
  description = "A map that contains the name and arn of event buses to forward events to"
}

variable "iam_role" {
  type        = any
  description = "IAM role to allow cross account put to event bus"
}

variable "receive_account_ids" {
  type        = list(string)
  description = "IDs of accounts to receive messages from"
  default     = []
}
