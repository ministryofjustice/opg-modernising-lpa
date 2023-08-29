variable "target_event_bus_arn" {
  type        = string
  description = "ARN of the event bus to forward events to"
}

variable "iam_role" {
  type        = any
  description = "IAM role to allow cross account put to event bus"
}
