resource "aws_iam_role" "execution_role" {
  name               = "${local.environment_name}-execution-role-ecs-cluster"
  assume_role_policy = data.aws_iam_policy_document.execution_role_assume_policy.json
  provider           = aws.global
}

data "aws_iam_policy_document" "execution_role_assume_policy" {
  statement {
    effect  = "Allow"
    actions = ["sts:AssumeRole"]

    principals {
      identifiers = ["ecs-tasks.amazonaws.com"]
      type        = "Service"
    }
  }
  provider = aws.global
}

resource "aws_iam_role_policy" "execution_role" {
  name     = "${local.environment_name}-execution-role"
  policy   = data.aws_iam_policy_document.execution_role.json
  role     = aws_iam_role.execution_role.id
  provider = aws.global
}

data "aws_secretsmanager_secret" "rum_monitor_identity_pool_id_eu_west_1" {
  count    = local.environment.app.rum_enabled ? 1 : 0
  name     = "rum-monitor-identity-pool-id"
  provider = aws.eu_west_1
}

data "aws_secretsmanager_secret" "rum_monitor_identity_pool_id_eu_west_2" {
  count    = local.environment.app.rum_enabled ? 1 : 0
  name     = "rum-monitor-identity-pool-id"
  provider = aws.eu_west_2
}


data "aws_iam_policy_document" "execution_role" {
  statement {
    effect    = "Allow"
    resources = ["*"] #tfsec:ignore:aws-iam-no-policy-wildcards
    actions = [
      "ecr:GetAuthorizationToken",
      "ecr:BatchCheckLayerAvailability",
      "ecr:GetDownloadUrlForLayer",
      "ecr:BatchGetImage",
    ]
  }
  statement {
    effect    = "Allow"
    resources = ["*"] #tfsec:ignore:aws-iam-no-policy-wildcards
    actions = [
      "logs:CreateLogStream",
      "logs:PutLogEvents",
    ]
  }
  statement {
    effect = "Allow"

    resources = [
      data.aws_secretsmanager_secret.rum_monitor_identity_pool_id_eu_west_1[0].arn,
      data.aws_secretsmanager_secret.rum_monitor_identity_pool_id_eu_west_2[0].arn,
      aws_secretsmanager_secret.rum_monitor_application_id[0].arn,
    ]

    actions = [
      "secretsmanager:GetSecretValue",
    ]
  }
  statement {
    effect = "Allow"

    resources = [
      data.aws_kms_alias.secrets_manager_secret_encryption_key_eu_west_1[0].target_key_arn,
      data.aws_kms_alias.secrets_manager_secret_encryption_key_eu_west_2[0].target_key_arn,
    ]

    actions = [
      "kms:Decrypt",
      "kms:GenerateDataKey",
      "kms:GenerateDataKeyPair",
      "kms:GenerateDataKeyPairWithoutPlaintext",
      "kms:GenerateDataKeyWithoutPlaintext",
      "kms:DescribeKey",
    ]
  }
  provider = aws.global
}
