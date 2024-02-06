# Create encrypted logging for fault injection experiments

data "aws_kms_alias" "cloudwatch_application_logs_encryption" {
  name     = "alias/${data.aws_default_tags.current.tags.application}_cloudwatch_application_logs_encryption"
  provider = aws.region
}

resource "aws_cloudwatch_log_group" "fis_app_ecs_tasks" {
  name              = "/aws/fis/app-ecs-tasks-experiment-${data.aws_default_tags.current.tags.environment-name}"
  retention_in_days = 7
  kms_key_id        = data.aws_kms_alias.cloudwatch_application_logs_encryption.target_key_arn
  provider          = aws.region
}

# Add resource policy to allow FIS or the FIS role to write logs - not working

data "aws_iam_policy_document" "cloudwatch_log_group_policy_fis_app_ecs_tasks" {
  provider = aws.region
  statement {
    sid    = "AWSLogDeliveryWrite20150319"
    effect = "Allow"

    principals {
      identifiers = [
        "delivery.logs.amazonaws.com"
      ]
      type = "Service"
    }

    actions = [
      "logs:CreateLogStream",
      "logs:PutLogEvents"
    ]

    resources = [
      "${aws_cloudwatch_log_group.fis_app_ecs_tasks.arn}:*",
    ]

    condition {
      test     = "StringEquals"
      variable = "aws:SourceAccount"
      values   = [data.aws_caller_identity.current.account_id]
    }
  }
}

#  Add resource policy to allow FIS or the FIS role to write logs - not working
resource "aws_cloudwatch_log_resource_policy" "fis_app_ecs_tasks" {
  provider        = aws.region
  policy_document = data.aws_iam_policy_document.cloudwatch_log_group_policy_fis_app_ecs_tasks.json
  policy_name     = "fis_app_ecs_tasks_logging"
}

# Add log encryption and log write/delivery permissions to the FIS role

data "aws_iam_policy_document" "fis_role_log_encryption" {
  provider  = aws.region
  policy_id = "log_access"
  statement {
    sid = "AllowCloudWatchLogsEncryption"
    actions = [
      "kms:Encrypt",
      "kms:GenerateDataKey",
    ]

    resources = [
      data.aws_kms_alias.cloudwatch_application_logs_encryption.target_key_arn
    ]
  }

  statement {
    sid    = "AllowCloudWatchLogs"
    effect = "Allow"
    actions = [
      "logs:CreateLogDelivery",
      "logs:DescribeLogGroups",
      "logs:CreateLogStream",
      "logs:PutLogEvents",
      "logs:DescribeResourcePolicies",
    ]

    resources = [
      aws_cloudwatch_log_group.fis_app_ecs_tasks.arn,
      "${aws_cloudwatch_log_group.fis_app_ecs_tasks.arn}:*",
    ]
  }

  # statement {
  #   sid    = "AllowFISExperimentRoleCloudwatch"
  #   effect = "Allow"
  #   actions = [
  #     "logs:Describe*",
  #     "logs:CreateLogDelivery",
  #     "logs:PutLogEvents",
  #     "logs:CreateLogStream",
  #     "logs:PutResourcePolicy"
  #   ]

  #   resources = [
  #     "*"
  #   ]
  # }

  # statement {
  #   sid    = "AllowLogDeliveryActions"
  #   effect = "Allow"
  #   actions = [
  #     "logs:PutDeliverySource",
  #     "logs:GetDeliverySource",
  #     "logs:DeleteDeliverySource",
  #     "logs:DescribeDeliverySources",
  #     "logs:PutDeliveryDestination",
  #     "logs:GetDeliveryDestination",
  #     "logs:DeleteDeliveryDestination",
  #     "logs:DescribeDeliveryDestinations",
  #     "logs:CreateDelivery",
  #     "logs:GetDelivery",
  #     "logs:DeleteDelivery",
  #     "logs:DescribeDeliveries",
  #     "logs:PutDeliveryDestinationPolicy",
  #     "logs:GetDeliveryDestinationPolicy",
  #     "logs:DeleteDeliveryDestinationPolicy"
  #   ]
  #   resources = [
  #     "arn:aws:logs:${data.aws_region.current.name}:${data.aws_caller_identity.current.account_id}:delivery-source:*",
  #     "arn:aws:logs:${data.aws_region.current.name}:${data.aws_caller_identity.current.account_id}:delivery:*",
  #     "arn:aws:logs:${data.aws_region.current.name}:${data.aws_caller_identity.current.account_id}:delivery-destination:*"
  #   ]
  # }

  # statement {
  #   sid    = "AllowUpdatesToResourcePolicyCWL"
  #   effect = "Allow"
  #   actions = [
  #     "logs:PutResourcePolicy",
  #     "logs:DescribeResourcePolicies",
  #     "logs:DescribeLogGroups"
  #   ]
  #   resources = [
  #     aws_cloudwatch_log_group.fis_app_ecs_tasks.arn,
  #     "${aws_cloudwatch_log_group.fis_app_ecs_tasks.arn}:*",
  #   ]
  # }
}

resource "aws_iam_role_policy" "fis_role_log_encryption" {
  provider = aws.region
  name     = "fis-role-log-permissions"
  role     = var.fault_injection_simulator_role.name
  policy   = data.aws_iam_policy_document.fis_role_log_encryption.json
}

# Create experiment template for ECS tasks

resource "aws_fis_experiment_template" "ecs_app" {
  provider    = aws.region
  description = "Run ECS task experiments for the app service"
  role_arn    = var.fault_injection_simulator_role.arn
  tags = {
    Name = "${data.aws_default_tags.current.tags.environment-name} - APP ECS Task Experiments"
  }

  action {
    action_id   = "aws:ecs:task-cpu-stress"
    description = null
    name        = "cpu_stress_100_percent"
    parameter {
      key   = "duration"
      value = "PT5M"
    }
    target {
      key   = "Tasks"
      value = "app-ecs-tasks-${data.aws_default_tags.current.tags.environment-name}"
    }
  }

  stop_condition {
    source = "none"
    value  = null
  }

  log_configuration {
    log_schema_version = 2

    cloudwatch_logs_configuration {
      log_group_arn = "${aws_cloudwatch_log_group.fis_app_ecs_tasks.arn}:*" # tfsec:ignore:aws-cloudwatch-log-group-wildcard
    }
  }

  target {
    name = "app-ecs-tasks-${data.aws_default_tags.current.tags.environment-name}"
    resource_tag {
      key   = "environment-name"
      value = data.aws_default_tags.current.tags.environment-name
    }
    resource_type  = "aws:ecs:task"
    selection_mode = "ALL"
  }
}
