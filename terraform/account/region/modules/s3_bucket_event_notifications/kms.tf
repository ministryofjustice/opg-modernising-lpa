resource "aws_kms_key" "s3_bucket_event_notification_sns" {
  description             = "KMS Key for Encrypting S3 Bucket Event Notifications in SNS"
  deletion_window_in_days = 10
  policy                  = data.aws_iam_policy_document.s3_bucket_event_notification_sns.json
  enable_key_rotation     = true
}

data "aws_iam_policy_document" "s3_bucket_event_notification_sns" {
  statement {
    sid       = "Allow Key to be used for Encryption by AWS Config Role"
    effect    = "Allow"
    resources = ["*"]
    actions = [
      "kms:Decrypt",
      "kms:GenerateDataKey*",
    ]
    principals {
      type        = "Service"
      identifiers = ["s3.amazonaws.com"]
    }
  }
  statement {
    sid       = "Enable Root account permissions on Key"
    effect    = "Allow"
    actions   = ["kms:*"]
    resources = ["*"]

    principals {
      type = "AWS"
      identifiers = [
        "arn:aws:iam::${data.aws_caller_identity.current.account_id}:root",
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
      "kms:CancelKeyDeletion"
    ]

    principals {
      type        = "AWS"
      identifiers = ["arn:aws:iam::${data.aws_caller_identity.current.account_id}:role/breakglass"]
    }
  }
}
