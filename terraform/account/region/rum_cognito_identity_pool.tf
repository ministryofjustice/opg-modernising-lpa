resource "aws_iam_role" "rum_monitor_unauthenticated" {
  count              = var.rum_enabled ? 1 : 0
  name               = "RUM-Monitor-Unauthenticated-${data.aws_region.current.name}"
  assume_role_policy = data.aws_iam_policy_document.rum_monitor_unauthenticated_role_assume_policy[0].json
  provider           = aws.global
}

data "aws_iam_policy_document" "rum_monitor_unauthenticated_role_assume_policy" {
  count = var.rum_enabled ? 1 : 0
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
  provider = aws.global
}

resource "aws_cognito_identity_pool" "rum_monitor" {
  count                            = var.rum_enabled ? 1 : 0
  identity_pool_name               = "RUM-Monitor-${data.aws_region.current.name}"
  allow_unauthenticated_identities = true
  allow_classic_flow               = true
  provider                         = aws.region
}

resource "aws_cognito_identity_pool_roles_attachment" "rum_monitor" {
  count            = var.rum_enabled ? 1 : 0
  identity_pool_id = aws_cognito_identity_pool.rum_monitor[0].id
  roles = {
    unauthenticated = aws_iam_role.rum_monitor_unauthenticated[0].arn
  }
  provider = aws.region
}

resource "aws_secretsmanager_secret" "rum_monitor_identity_pool_id" {
  count                   = var.rum_enabled ? 1 : 0
  name                    = "rum-monitor-identity-pool-id-${data.aws_region.current.name}"
  kms_key_id              = data.aws_kms_alias.secrets_manager.target_key_id
  recovery_window_in_days = 0
  provider                = aws.region
}

resource "aws_secretsmanager_secret_version" "rum_monitor_identity_pool_id" {
  count         = var.rum_enabled ? 1 : 0
  secret_id     = aws_secretsmanager_secret.rum_monitor_identity_pool_id[0].id
  secret_string = aws_cognito_identity_pool.rum_monitor[0].id
  provider      = aws.region
}
