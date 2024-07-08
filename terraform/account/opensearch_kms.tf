module "opensearch_kms" {
  source                  = "./modules/kms_key"
  encrypted_resource      = "opensearch"
  kms_key_alias_name      = "${local.default_tags.application}-opensearch-encryption-key"
  enable_key_rotation     = true
  enable_multi_region     = true
  deletion_window_in_days = 10
  kms_key_policy          = local.account.account_name == "development" ? data.aws_iam_policy_document.opensearch_kms_merged.json : data.aws_iam_policy_document.opensearch_kms.json
  providers = {
    aws.eu_west_1 = aws.eu_west_1
    aws.eu_west_2 = aws.eu_west_2
  }
}

# See the following link for further information
# https://docs.aws.amazon.com/kms/latest/developerguide/key-policies.html
data "aws_iam_policy_document" "opensearch_kms_merged" {
  provider = aws.global
  source_policy_documents = [
    data.aws_iam_policy_document.opensearch_kms.json,
    data.aws_iam_policy_document.opensearch_kms_development_account_operator_admin.json
  ]
}

data "aws_iam_policy_document" "opensearch_kms" {
  provider = aws.global
  statement {
    sid    = "Allow Grant for Opensearch Serverless"
    effect = "Allow"
    resources = [
      "arn:aws:kms:*:${data.aws_caller_identity.global.account_id}:key/*"
    ]
    actions = [
      "kms:CreateGrant",
      "kms:DescribeKey",
    ]
    principals {
      type        = "AWS"
      identifiers = ["arn:aws:iam::${data.aws_caller_identity.global.account_id}:root"]
    }
    condition {
      test     = "Bool"
      variable = "kms:GrantIsForAWSResource"
      values   = ["true"]
    }
    condition {
      test     = "StringEquals"
      variable = "kms:ViaService"
      values = [
        "aoss.eu-west-1.amazonaws.com",
        "aoss.eu-west-2.amazonaws.com"
      ]
    }
  }

  statement {
    sid    = "Allow list keys for Opensearch Serverless"
    effect = "Allow"
    resources = [
      "*"
    ]
    actions = [
      "kms:ListKeys",
    ]
    principals {
      type        = "AWS"
      identifiers = ["arn:aws:iam::${data.aws_caller_identity.global.account_id}:root"]
    }
    condition {
      test     = "Bool"
      variable = "kms:GrantIsForAWSResource"
      values   = ["true"]
    }
    condition {
      test     = "StringEquals"
      variable = "kms:ViaService"
      values = [
        "aoss.eu-west-1.amazonaws.com",
        "aoss.eu-west-2.amazonaws.com"
      ]
    }
  }

  statement {
    sid    = "General View Access"
    effect = "Allow"
    resources = [
      "arn:aws:kms:*:${data.aws_caller_identity.global.account_id}:key/*"
    ]
    actions = [
      "kms:DescribeKey",
      "kms:GetKeyPolicy",
      "kms:GetKeyRotationStatus",
      "kms:List*",
    ]

    principals {
      type = "AWS"
      identifiers = [
        "arn:aws:iam::${data.aws_caller_identity.global.account_id}:root"
      ]
    }
  }

  statement {
    sid    = "Key Administrator"
    effect = "Allow"
    resources = [
      "arn:aws:kms:*:${data.aws_caller_identity.global.account_id}:key/*"
    ]
    actions = [
      "kms:Create*",
      "kms:Describe*",
      "kms:Enable*",
      "kms:List*",
      "kms:Put*",
      "kms:Update*",
      "kms:Revoke*",
      "kms:Disable*",
      "kms:Get*",
      "kms:Delete*",
      "kms:TagResource",
      "kms:UntagResource",
      "kms:ScheduleKeyDeletion",
      "kms:CancelKeyDeletion",
      "kms:ReplicateKey"
    ]

    principals {
      type = "AWS"
      identifiers = [
        "arn:aws:iam::${data.aws_caller_identity.global.account_id}:role/breakglass",
        "arn:aws:iam::${data.aws_caller_identity.global.account_id}:role/modernising-lpa-ci",
      ]
    }
  }
}

data "aws_iam_policy_document" "opensearch_kms_development_account_operator_admin" {
  provider = aws.global
  statement {
    sid    = "Dev Account Key Administrator"
    effect = "Allow"
    resources = [
      "arn:aws:kms:*:${data.aws_caller_identity.global.account_id}:key/*"
    ]
    actions = [
      "kms:Create*",
      "kms:Describe*",
      "kms:Enable*",
      "kms:List*",
      "kms:Put*",
      "kms:Update*",
      "kms:Revoke*",
      "kms:Disable*",
      "kms:Get*",
      "kms:Delete*",
      "kms:TagResource",
      "kms:UntagResource",
      "kms:ScheduleKeyDeletion",
      "kms:CancelKeyDeletion"
    ]

    principals {
      type = "AWS"
      identifiers = [
        "arn:aws:iam::${data.aws_caller_identity.global.account_id}:role/operator"
      ]
    }
  }
}
