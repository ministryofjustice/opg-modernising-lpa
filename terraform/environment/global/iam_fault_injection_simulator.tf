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

resource "aws_iam_role_policy_attachment" "fault_injection_simulator" {
  role       = aws_iam_role.fault_injection_simulator.name
  policy_arn = "arn:aws:iam::aws:policy/service-role/AWSFaultInjectionSimulatorECSAccess"
  provider   = aws.global
}

resource "aws_iam_role_policy" "fault_injection_simulator_additional_permissions" {
  name     = "additional-permissions"
  role     = aws_iam_role.fault_injection_simulator.name
  policy   = data.aws_iam_policy_document.fault_injection_simulator_combined.json
  provider = aws.global
}

#TODO: consolidate this into a single document if possible
data "aws_iam_policy_document" "fault_injection_simulator_combined" {
  source_policy_documents = [
    data.aws_iam_policy_document.fault_injection_simulator_additional_permissions.json,
    data.aws_iam_policy_document.fis_autocreated_role.json
  ]
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
    ]
  }
  statement {
    sid       = "AllowSSMCommands"
    effect    = "Allow"
    resources = ["*"] #tfsec:ignore:aws-iam-no-policy-wildcards
    actions = [
      "ssm:SendCommand",
      "ssm:ListCommands",
      "ssm:CancelCommand",
    ]
  }
}

data "aws_iam_policy_document" "fis_autocreated_role" {
  # example taken from role created in console to be attached to the FIS role
  version = "2012-10-17"

  statement {
    effect = "Allow"
    actions = [
      "ecs:DescribeClusters",
      "ecs:ListContainerInstances"
    ]
    resources = [
      "arn:aws:ecs:*:*:cluster/*"
    ]
  }

  statement {
    effect = "Allow"
    actions = [
      "ecs:DescribeTasks",
      "ecs:StopTask"
    ]
    resources = [
      "arn:aws:ecs:*:*:task/*/*"
    ]
  }

  statement {
    effect = "Allow"
    actions = [
      "ecs:ListTasks",
      "ecs:UpdateContainerInstancesState"
    ]
    resources = [
      "arn:aws:ecs:*:*:container-instance/*/*"
    ]
  }

  statement {
    effect = "Allow"
    actions = [
      "ssm:SendCommand"
    ]
    resources = [
      "arn:aws:ssm:*:*:managed-instance/*",
      "arn:aws:ssm:*:*:document/*"
    ]
  }

  #duplicate of above
  # statement {
  #   effect = "Allow"
  #   actions = [
  #     "ssm:ListCommands",
  #     "ssm:CancelCommand"
  #   ]
  #   resources = ["*"]
  # }
}
