resource "aws_iam_role_policy" "execution_role_region" {
  name     = "${data.aws_default_tags.current.tags.environment-name}-execution-role-${data.aws_region.current.region}"
  policy   = data.aws_iam_policy_document.execution_role_region.json
  role     = var.iam_roles.ecs_execution_role.id
  provider = aws.global
}

data "aws_kms_alias" "secrets_manager_secret_encryption_key" {
  name     = "alias/${data.aws_default_tags.current.tags.application}_secrets_manager_secret_encryption_key"
  provider = aws.region
}

data "aws_secretsmanager_secret" "gov_one_login_mrlpa_client_id" {
  name     = "gov-one-login-mrlpa-client-id"
  provider = aws.region
}

data "aws_secretsmanager_secret_version" "gov_one_login_mrlpa_client_id" {
  secret_id = data.aws_secretsmanager_secret.gov_one_login_mrlpa_client_id.id
  provider  = aws.region
}

data "aws_iam_policy_document" "execution_role_region" {
  statement {
    effect = "Allow"

    resources = [
      data.aws_secretsmanager_secret_version.rum_monitor_identity_pool_id.arn,
      data.aws_secretsmanager_secret_version.gov_one_login_mrlpa_client_id.arn,
      aws_secretsmanager_secret.rum_monitor_application_id.arn,
    ]

    actions = [
      "secretsmanager:GetSecretValue",
    ]
  }

  statement {
    effect = "Allow"

    resources = [
      data.aws_kms_alias.secrets_manager_secret_encryption_key.target_key_arn,
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
