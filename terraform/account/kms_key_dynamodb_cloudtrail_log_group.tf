
module "dynamodb_cloudtrail_log_group" {
  source                  = "./modules/kms_key"
  encrypted_resource      = "dynamodb cloudtrail log group"
  kms_key_alias_name      = "${local.default_tags.application}_dynamodb_cloudtrail_log_group_encryption"
  enable_key_rotation     = true
  enable_multi_region     = true
  deletion_window_in_days = 10
  kms_key_policy          = local.account.account_name == "development" ? data.aws_iam_policy_document.dynamodb_cloudtrail_log_group_merged.json : data.aws_iam_policy_document.dynamodb_cloudtrail_log_group.json
  providers = {
    aws.eu_west_1 = aws.eu_west_1
    aws.eu_west_2 = aws.eu_west_2
  }
}

# See the following link for further information
# https://docs.aws.amazon.com/kms/latest/developerguide/key-policies.html
data "aws_iam_policy_document" "dynamodb_cloudtrail_log_group_merged" {
  provider = aws.global
  source_policy_documents = [
    data.aws_iam_policy_document.dynamodb_cloudtrail_log_group.json,
    data.aws_iam_policy_document.dynamodb_cloudtrail_log_group_development_account_operator_admin.json
  ]
}

data "aws_iam_policy_document" "dynamodb_cloudtrail_log_group" {
  provider = aws.global

  statement {
    sid    = "Enable IAM User Permissions"
    effect = "Allow"
    principals {
      type        = "AWS"
      identifiers = ["arn:aws:iam::${data.aws_caller_identity.global.account_id}:root"]
    }
    actions = [
      "kms:*",
    ]
    resources = [
      "*",
    ]
  }

  statement {
    sid    = "Allow Key to be used for Encryption"
    effect = "Allow"
    resources = [
      "arn:aws:kms:*:${data.aws_caller_identity.global.account_id}:key/*"
    ]
    actions = [
      "kms:Encrypt",
      "kms:Decrypt",
      "kms:ReEncrypt*",
      "kms:GenerateDataKey*",
      "kms:DescribeKey",
    ]

    principals {
      type = "Service"
      identifiers = [
        "logs.${data.aws_region.eu_west_1.region}.amazonaws.com",
        "logs.${data.aws_region.eu_west_2.region}.amazonaws.com",
        "cloudtrail.amazonaws.com"
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

  statement {
    sid    = "DataAccessDecrypt"
    effect = "Allow"
    resources = [
      "arn:aws:kms:*:${data.aws_caller_identity.global.account_id}:key/*"
    ]
    actions = [
      "kms:Decrypt",
      "kms:GenerateDataKey*",
      "kms:DescribeKey",
    ]

    principals {
      type = "AWS"
      identifiers = [
        "arn:aws:iam::${data.aws_caller_identity.global.account_id}:role/data-access",
      ]
    }
  }
}

data "aws_iam_policy_document" "dynamodb_cloudtrail_log_group_development_account_operator_admin" {
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
