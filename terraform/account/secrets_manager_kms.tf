resource "aws_kms_key" "secrets_manager" {
  description             = "${local.mandatory_moj_tags.application} Secrets Manager secret encryption key"
  deletion_window_in_days = 10
  enable_key_rotation     = true
  policy                  = local.account.account_name == "development" ? data.aws_iam_policy_document.secrets_manager_kms_merged.json : data.aws_iam_policy_document.secrets_manager_kms.json
  multi_region            = true
  provider                = aws.eu_west_1
}

resource "aws_kms_replica_key" "secrets_manager_replica" {
  description             = "${local.mandatory_moj_tags.application} Secrets Manager secret multi-region replica key"
  deletion_window_in_days = 7
  primary_key_arn         = aws_kms_key.secrets_manager.arn
  provider                = aws.eu_west_2
}

resource "aws_kms_alias" "secrets_manager_alias_eu_west_1" {
  name          = "alias/${local.mandatory_moj_tags.application}_secrets_manager_secret_encryption_key"
  target_key_id = aws_kms_key.secrets_manager.key_id
  provider      = aws.eu_west_1
}

resource "aws_kms_alias" "secrets_manager_alias_eu_west_2" {
  name          = "alias/${local.mandatory_moj_tags.application}_secrets_manager_secret_encryption_key"
  target_key_id = aws_kms_replica_key.secrets_manager_replica.key_id
  provider      = aws.eu_west_2
}

# See the following link for further information
# https://docs.aws.amazon.com/kms/latest/developerguide/key-policies.html
data "aws_iam_policy_document" "secrets_manager_kms_merged" {
  provider = aws.global
  source_policy_documents = [
    data.aws_iam_policy_document.secrets_manager_kms.json,
    data.aws_iam_policy_document.secrets_manager_kms_development_account_operator_admin.json
  ]
}

data "aws_iam_policy_document" "secrets_manager_kms" {
  provider = aws.global
  statement {
    sid       = "Allow Key to be used for Encryption"
    effect    = "Allow"
    resources = ["*"]
    actions = [
      "kms:Encrypt",
      "kms:ReEncrypt*",
      "kms:GenerateDataKey*",
      "kms:DescribeKey",
    ]

    principals {
      type = "AWS"
      identifiers = [
        "arn:aws:iam::${data.aws_caller_identity.global.account_id}:role/breakglass",
      ]
    }
  }

  statement {
    sid       = "Allow Key to be used for Decryption"
    effect    = "Allow"
    resources = ["*"]
    actions = [
      "kms:Decrypt",
      "kms:GenerateDataKey*",
      "kms:DescribeKey",
    ]

    principals {
      type = "AWS"
      identifiers = [
        # need a better principle and condition as per
        #  https://docs.aws.amazon.com/secretsmanager/latest/userguide/security-encryption.html
        "arn:aws:iam::${data.aws_caller_identity.global.account_id}:root",
        # ECS task role for app
      ]
    }
  }

  statement {
    sid       = "Key Administrator"
    effect    = "Allow"
    resources = ["*"]
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

data "aws_iam_policy_document" "secrets_manager_kms_development_account_operator_admin" {
  provider = aws.global
  statement {
    sid       = "Dev Account Key Administrator"
    effect    = "Allow"
    resources = ["*"]
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
