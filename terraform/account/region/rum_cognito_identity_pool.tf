resource "aws_iam_role" "rum_monitor_unauthenticated" {
  name               = "RUM-Monitor-Unauthenticated-${data.aws_region.current.name}"
  assume_role_policy = data.aws_iam_policy_document.rum_monitor_unauthenticated_role_assume_policy.json
  provider           = aws.global
}

data "aws_iam_policy_document" "rum_monitor_unauthenticated_role_assume_policy" {
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
        aws_cognito_identity_pool.rum_monitor.id
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
  identity_pool_name               = "RUM-Monitor-${data.aws_region.current.name}"
  allow_unauthenticated_identities = true
  allow_classic_flow               = true
  provider                         = aws.region
}

resource "aws_cognito_identity_pool_roles_attachment" "rum_monitor" {
  identity_pool_id = aws_cognito_identity_pool.rum_monitor.id
  roles = {
    unauthenticated = aws_iam_role.rum_monitor_unauthenticated.arn
  }
  provider = aws.region
}

resource "aws_secretsmanager_secret" "rum_monitor_identity_pool_id" {
  name                    = "rum-monitor-identity-pool-id-${data.aws_region.current.name}"
  kms_key_id              = data.aws_kms_alias.secrets_manager.target_key_id
  recovery_window_in_days = 0
  provider                = aws.region
}

resource "aws_secretsmanager_secret_version" "rum_monitor_identity_pool_id" {
  secret_id     = aws_secretsmanager_secret.rum_monitor_identity_pool_id.id
  secret_string = aws_cognito_identity_pool.rum_monitor.id
  provider      = aws.region
}
