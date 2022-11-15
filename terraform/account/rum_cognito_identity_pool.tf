resource "aws_cognito_identity_pool" "rum_monitor" {
  identity_pool_name               = "RUM-Monitor-${data.aws_region.eu_west_1.name}"
  allow_unauthenticated_identities = true
  provider                         = aws.eu_west_1
}

resource "aws_iam_role" "rum_monitor_unauthenticated" {
  name               = "RUM-Monitor-${data.aws_region.eu_west_1.name}"
  assume_role_policy = data.aws_iam_policy_document.rum_monitor_unauthenticated_role_assume_policy.json
  provider           = aws.eu_west_1
}


data "aws_iam_policy_document" "rum_monitor_unauthenticated_role_assume_policy" {
  statement {
    effect  = "Allow"
    actions = ["sts:AssumeRoleWithWebIdentity"]

    principals {
      identifiers = ["cognito-identity.amazonaws.com"]
      type        = "Federated"
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

      values = ["unauthenticated"]
    }
  }
  provider = aws.eu_west_1
}
