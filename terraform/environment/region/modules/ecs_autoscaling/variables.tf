variable "aws_ecs_cluster_name" {
  description = "Name of the ECS cluster for the service being scaled."
  type        = string
}

variable "aws_ecs_service_name" {
  description = "Name of the ECS service."
  type        = string
}

variable "max_scaling_alarm_actions" {
  description = "List of alarm actions for maximum autoscaling being reached."
  type        = list(string)
}

variable "ecs_autoscaling_service_role_arn" {
  description = "The ARN of the IAM role that allows Application AutoScaling to modify your scalable target on your behalf."
  type        = string
}

variable "environment_name" {
  description = "Name of the environment."
  type        = string
}

variable "region_name" {
  description = "region name"
  type        = string
}

variable "ecs_task_autoscaling_maximum" {
  description = "The max capacity of the scalable target."
  type        = number
}

variable "autoscaling_metric_max_cpu_target" {
  description = "The target value for the CPU metric."
  type        = number
  default     = 80
}

variable "autoscaling_metric_max_memory_target" {
  description = "The target value for the memory metric."
  type        = number
  default     = 80
}

variable "autoscaling_metric_min_cpu_target" {
  description = "The target value for the CPU metric."
  type        = number
  default     = 30
}

variable "autoscaling_metric_min_memory_target" {
  description = "The target value for the memory metric."
  type        = number
  default     = 30
}

variable "ecs_task_autoscaling_minimum" {
  description = "The min capacity of the scalable target."
  type        = number
  default     = 1
}

variable "scale_down_cooldown" {
  description = "The amount of time, in seconds, after a scale in activity completes before another scale in activity can start."
  type        = number
  default     = 60
}

variable "scale_up_cooldown" {
  description = "The amount of time, in seconds, after a scale out activity completes before another scale out activity can start."
  type        = number
  default     = 60
}
