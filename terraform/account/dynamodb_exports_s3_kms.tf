resource "aws_kms_key" "dynamodb_exports_s3_bucket" {
  description             = "${local.default_tags.application} dynamodb exports s3 bucket encryption key"
  deletion_window_in_days = 10
  enable_key_rotation     = true
  policy                  = local.account.account_name == "development" ? data.aws_iam_policy_document.dynamodb_exports_s3_bucket_kms_merged.json : data.aws_iam_policy_document.dynamodb_exports_s3_bucket_kms.json
  multi_region            = true
  provider                = aws.eu_west_1
  lifecycle {
    prevent_destroy = true
  }
}

resource "aws_kms_replica_key" "dynamodb_exports_s3_bucket_replica" {
  description             = "${local.default_tags.application} dynamodb exprots s3 Multi-Region replica key"
  deletion_window_in_days = 7
  primary_key_arn         = aws_kms_key.dynamodb_exports_s3_bucket.arn
  provider                = aws.eu_west_2
  lifecycle {
    prevent_destroy = true
  }
}

resource "aws_kms_alias" "dynamodb_exports_s3_bucket_alias_eu_west_1" {
  name          = "alias/${local.default_tags.application}-dynamodb-exports-s3-bucket-encryption"
  target_key_id = aws_kms_key.dynamodb_exports_s3_bucket.key_id
  provider      = aws.eu_west_1
}

resource "aws_kms_alias" "dynamodb_exports_s3_bucket_alias_eu_west_2" {
  name          = "alias/${local.default_tags.application}-dynamodb-exports-s3-bucket-encryption"
  target_key_id = aws_kms_replica_key.dynamodb_exports_s3_bucket_replica.key_id
  provider      = aws.eu_west_2
}

# See the following link for further information
# https://docs.aws.amazon.com/kms/latest/developerguide/key-policies.html
data "aws_iam_policy_document" "dynamodb_exports_s3_bucket_kms_merged" {
  provider = aws.global
  source_policy_documents = [
    data.aws_iam_policy_document.dynamodb_exports_s3_bucket_kms.json,
    data.aws_iam_policy_document.dynamodb_exports_s3_bucket_kms_development_account_operator_admin.json
  ]
}

data "aws_iam_policy_document" "dynamodb_exports_s3_bucket_kms" {
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
      type = "AWS"
      identifiers = [
        aws_iam_role.opensearch_ingestion_service_role.arn
      ]
    }
    condition {
      test     = "StringLike"
      variable = "kms:ViaService"

      values = [
        "s3.*.amazonaws.com"
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
      "kms:ReplicateKey",
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
    sid    = "Key Administrator Decryption"
    effect = "Allow"
    resources = [
      "arn:aws:kms:*:${data.aws_caller_identity.global.account_id}:key/*"
    ]
    actions = [
      "kms:Decrypt",
    ]

    principals {
      type = "AWS"
      identifiers = [
        "arn:aws:iam::${data.aws_caller_identity.global.account_id}:role/breakglass",
      ]
    }
  }
}

data "aws_iam_policy_document" "dynamodb_exports_s3_bucket_kms_development_account_operator_admin" {
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
      "kms:CancelKeyDeletion",
    ]

    principals {
      type = "AWS"
      identifiers = [
        "arn:aws:iam::${data.aws_caller_identity.global.account_id}:role/operator"
      ]
    }
  }
}
