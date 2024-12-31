variable "iam_roles" {
  type = object({
    ecs_execution_role                      = any
    app_ecs_task_role                       = any
    s3_antivirus                            = any
    cross_account_put                       = any
    fault_injection_simulator               = any
    create_s3_batch_replication_jobs_lambda = any
    event_received_lambda                   = any
    schedule_runner_scheduler               = any
    schedule_runner_lambda                  = any
  })
  description = "ARN of IAM role that allows your Amazon ECS container task to make calls to other AWS services."
}

variable "application_log_retention_days" {
  type        = number
  description = "Specifies the number of days you want to retain log events in the specified log group. Possible values are: 1, 3, 5, 7, 14, 30, 60, 90, 120, 150, 180, 365, 400, 545, 731, 1827, 3653, and 0. If you select 0, the events in the log group are always retained and never expire."
}

variable "ecs_capacity_provider" {
  type        = string
  description = "Name of the capacity provider to use. Valid values are FARGATE_SPOT and FARGATE"
}

variable "ecs_task_autoscaling" {
  type        = any
  description = "task minimum and maximum values for autoscaling"
}

variable "app_service_repository_url" {
  type        = string
  description = "Repository URL for the app service"
}

variable "app_service_container_version" {
  type        = string
  description = "Container version the app service"
}

variable "mock_onelogin_service_repository_url" {
  type        = string
  description = "Repository URL for the mock-onelogin service"
}

variable "mock_onelogin_service_container_version" {
  type        = string
  description = "Container version for the mock-onelogin service"
}

variable "mock_pay_service_repository_url" {
  type        = string
  description = "Repository URL for the mock-pay service"
}

variable "mock_pay_service_container_version" {
  type        = string
  description = "Container version for the mock-pay service"
}

variable "ingress_allow_list_cidr" {
  type        = list(string)
  description = "List of CIDR ranges permitted to access the service"
}

variable "alb_deletion_protection_enabled" {
  type        = bool
  description = "If true, deletion of the load balancer will be disabled via the AWS API. This will prevent Terraform from deleting the load balancer. Defaults to false."
}

variable "lpas_table" {
  type        = any
  description = "DynamoDB table for storing LPAs"
}

variable "app_env_vars" {
  type        = any
  description = "Environment variable values for app"
}

variable "public_access_enabled" {
  type        = bool
  description = "Enable access to the Modernising LPA service from the public internet"
}

variable "pagerduty_service_name" {
  type        = string
  description = "Name of the PagerDuty service to use for alerts"
}

variable "dns_weighting" {
  type        = number
  description = "Weighting for DNS records"
}

variable "reduced_fees" {
  type = object({
    s3_object_replication_enabled             = bool
    target_environment                        = string
    destination_account_id                    = string
    enable_s3_batch_job_replication_scheduler = bool
  })
}

variable "target_event_bus_arn" {
  type        = string
  description = "ARN of the event bus to forward events to"
}

variable "receive_account_ids" {
  type        = list(string)
  description = "IDs of accounts to receive messages from"
  default     = []
}

variable "s3_antivirus_provisioned_concurrency" {
  type        = number
  description = "Number of concurrent executions to provision for Lambda"
  default     = 0
  validation {
    condition     = var.s3_antivirus_provisioned_concurrency >= 0 && var.s3_antivirus_provisioned_concurrency <= 6
    error_message = "s3_antivirus_provisioned_concurrency must be between 0 and 6"
  }
}

variable "uid_service" {
  type = object({
    base_url = string
    api_arns = list(string)
  })
}

variable "lpa_store_service" {
  type = object({
    base_url = string
    api_arns = list(string)
  })
}

variable "mock_onelogin_enabled" {
  type = bool
}

variable "mock_pay_enabled" {
  type = bool
}

variable "dependency_health_check_alarm_enabled" {
  type        = bool
  description = "Enable the dependency health check alert actions"
  default     = false
}

variable "service_health_check_alarm_enabled" {
  type        = bool
  description = "Enable the service health check alert actions"
  default     = false
}

variable "cloudwatch_application_insights_enabled" {
  type        = bool
  description = "Enable CloudWatch Application Insights"
}

variable "fault_injection_experiments_enabled" {
  type        = bool
  description = "Enable fault injection"
}

variable "search_endpoint" {
  type        = string
  description = "URL of the OpenSearch Service endpoint to use"
  nullable    = true
}

variable "search_index_name" {
  type        = string
  description = "Name of the OpenSearch Service index to use"
}

variable "search_collection_arn" {
  type        = string
  description = "ARN of the OpenSearch collection to use"
  nullable    = true
}

variable "real_user_monitoring_cw_logs_enabled" {
  type        = bool
  description = "Enable CloudWatch logging for Real User Monitoring"
}

variable "waf_alb_association_enabled" {
  type        = bool
  description = "Enable WAF association with the ALBs"
  default     = true
}

variable "egress_checker_repository_url" {
  type        = string
  description = "Repository URL for the egress-checker lambda function"
}

variable "egress_checker_container_version" {
  type        = string
  description = "Container version the egress-checker lambda function"
}
