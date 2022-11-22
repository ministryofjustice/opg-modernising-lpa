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
  name     = "rum-monitor-identity-pool-id-eu-west-1"
  provider = aws.eu_west_1
}

data "aws_kms_alias" "secrets_manager_secret_encryption_key_eu_west_1" {
  name     = "alias/${local.default_tags.application}_secrets_manager_secret_encryption_key"
  provider = aws.eu_west_1
}

data "aws_kms_alias" "secrets_manager_secret_encryption_key_eu_west_2" {
  name     = "alias/${local.default_tags.application}_secrets_manager_secret_encryption_key"
  provider = aws.eu_west_2
}

resource "aws_secretsmanager_secret" "rum_monitor_application_id_eu_west_1" {
  name       = "${local.environment_name}_rum_monitor_application_id"
  kms_key_id = data.aws_kms_alias.secrets_manager_secret_encryption_key_eu_west_1.target_key_id
  provider   = aws.eu_west_1
}

resource "aws_secretsmanager_secret" "rum_monitor_application_id_eu_west_2" {
  name       = "${local.environment_name}_rum_monitor_application_id"
  kms_key_id = data.aws_kms_alias.secrets_manager_secret_encryption_key_eu_west_1.target_key_id
  provider   = aws.eu_west_2
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
      data.aws_secretsmanager_secret.rum_monitor_identity_pool_id_eu_west_1.arn,
      aws_secretsmanager_secret.rum_monitor_application_id_eu_west_1.arn,
      aws_secretsmanager_secret.rum_monitor_application_id_eu_west_2.arn,
    ]

    actions = [
      "secretsmanager:GetSecretValue",
    ]
  }
  statement {
    effect = "Allow"

    resources = [
      data.aws_kms_alias.secrets_manager_secret_encryption_key_eu_west_1.target_key_arn,
      data.aws_kms_alias.secrets_manager_secret_encryption_key_eu_west_2.target_key_arn,
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
