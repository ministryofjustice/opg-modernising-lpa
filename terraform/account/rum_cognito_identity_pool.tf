resource "aws_cognito_identity_pool" "rum_monitor" {
  count                            = local.account.rum_enabled ? 1 : 0
  identity_pool_name               = "RUM-Monitor-${data.aws_region.eu_west_1.name}"
  allow_unauthenticated_identities = true
  provider                         = aws.eu_west_1
}

resource "aws_iam_role" "rum_monitor_unauthenticated" {
  count              = local.account.rum_enabled ? 1 : 0
  name               = "RUM-Monitor-Unauthenticated"
  assume_role_policy = data.aws_iam_policy_document.rum_monitor_unauthenticated_role_assume_policy[0].json
  provider           = aws.eu_west_1
}

data "aws_iam_policy_document" "rum_monitor_unauthenticated_role_assume_policy" {
  count = local.account.rum_enabled ? 1 : 0
  statement {
    effect  = "Allow"
    actions = ["sts:AssumeRoleWithWebIdentity"]

    principals {
      type        = "Federated"
      identifiers = ["cognito-identity.amazonaws.com"]
    }
    condition {
      test     = "StringEquals"
      variable = "cognito-identity.amazonaws.com:aud"

      values = [
        aws_cognito_identity_pool.rum_monitor[0].id
      ]
    }
    condition {
      test     = "ForAnyValue:StringLike"
      variable = "cognito-identity.amazonaws.com:amr"
      values   = ["unauthenticated"]
    }
  }
  provider = aws.eu_west_1
}

resource "aws_cognito_identity_pool_roles_attachment" "rum_monitor" {
  count            = local.account.rum_enabled ? 1 : 0
  identity_pool_id = aws_cognito_identity_pool.rum_monitor[0].id
  roles = {
    unauthenticated = aws_iam_role.rum_monitor_unauthenticated[0].arn
  }
  provider = aws.eu_west_1
}

resource "aws_ssm_parameter" "rum_monitor_identity_pool_id" {
  count    = local.account.rum_enabled ? 1 : 0
  name     = "rum_monitor_identity_pool_id"
  type     = "String"
  value    = aws_cognito_identity_pool.rum_monitor[0].id
  provider = aws.eu_west_1
}

resource "aws_secretsmanager_secret" "rum_monitor_identity_pool_id" {
  count                   = local.account.rum_enabled ? 1 : 0
  name                    = "rum-monitor-identity-pool-id"
  kms_key_id              = aws_kms_key.secrets_manager.key_id
  recovery_window_in_days = 0
  replica {
    kms_key_id = aws_kms_replica_key.secrets_manager_replica.key_id
    region     = data.aws_region.eu_west_2.name
  }
  provider = aws.eu_west_1
}

resource "aws_secretsmanager_secret_version" "rum_monitor_identity_pool_id" {
  count         = local.account.rum_enabled ? 1 : 0
  secret_id     = aws_secretsmanager_secret.rum_monitor_identity_pool_id[0].id
  secret_string = aws_cognito_identity_pool.rum_monitor[0].id
  provider      = aws.eu_west_1
}
