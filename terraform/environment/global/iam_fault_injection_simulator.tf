# Create role for running experiments

resource "aws_iam_role" "fault_injection_simulator" {
  name               = "fault-injection-simulator-${data.aws_default_tags.current.tags.environment-name}"
  assume_role_policy = data.aws_iam_policy_document.fault_injection_simulator_assume.json
  provider           = aws.global
}

data "aws_iam_policy_document" "fault_injection_simulator_assume" {
  statement {
    actions = ["sts:AssumeRole"]

    principals {
      type        = "Service"
      identifiers = ["fis.amazonaws.com"]
    }
  }
  provider = aws.global
}

# Add permissions for FIS to run experiments (ECS, Logging, SSM)

resource "aws_iam_role_policy_attachment" "fault_injection_simulator_ecs_access" {
  role       = aws_iam_role.fault_injection_simulator.name
  policy_arn = "arn:aws:iam::aws:policy/service-role/AWSFaultInjectionSimulatorECSAccess"
  provider   = aws.global
}

resource "aws_iam_role_policy_attachment" "fault_injection_simulator_ssm_access" {
  role       = aws_iam_role.fault_injection_simulator.name
  policy_arn = "arn:aws:iam::aws:policy/service-role/AWSFaultInjectionSimulatorSSMAccess"
  provider   = aws.global
}

resource "aws_iam_role_policy" "fault_injection_simulator_additional_permissions" {
  name     = "additional-permissions"
  role     = aws_iam_role.fault_injection_simulator.name
  policy   = data.aws_iam_policy_document.fault_injection_simulator_additional_permissions.json
  provider = aws.global
}

data "aws_iam_policy_document" "fault_injection_simulator_additional_permissions" {
  policy_id = "fix experiment permissions"
  statement {
    sid       = "AllowServiceLinkedRole"
    effect    = "Allow"
    resources = ["*"] #tfsec:ignore:aws-iam-no-policy-wildcards
    actions = [
      "iam:CreateServiceLinkedRole",
    ]
    condition {
      test     = "StringLike"
      variable = "iam:AWSServiceName"
      values   = ["fis.amazonaws.com"]
    }
    condition {
      test     = "StringEquals"
      variable = "aws:SourceAccount"
      values   = [data.aws_caller_identity.global.account_id]
    }
    condition {
      test     = "ArnLike"
      variable = "aws:SourceArn"
      values   = ["arn:aws:fis:${data.aws_region.global.name}:${data.aws_caller_identity.global.account_id}:experiment/*"]
    }
  }

  statement {
    sid       = "AllowCloudWatchLogs"
    effect    = "Allow"
    resources = ["*"] #tfsec:ignore:aws-iam-no-policy-wildcards
    actions = [
      "logs:CreateLogDelivery",
      "logs:DescribeLogGroups",
      "logs:CreateLogStream",
      "logs:PutLogEvents",
      "logs:DescribeResourcePolicies",
    ]
  }
}
