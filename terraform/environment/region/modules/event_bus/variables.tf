variable "target_event_bus_arns" {
  type        = map(string)
  description = "A map that contains the name and arn of event buses to forward events to"
}

variable "iam_role" {
  type        = any
  description = "IAM role to allow cross account put to event bus"
}

variable "opg_metrics_api_destination_role" {
  type        = any
  description = "IAM role to allow api destination calls to opg-metrics"
}

variable "receive_account_ids" {
  type        = list(string)
  description = "IDs of accounts to receive messages from"
  default     = []
}

variable "log_emitted_events" {
  type        = bool
  description = "Log events emitted to /aws/events/{env}-emitted"
  default     = false
}
