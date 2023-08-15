module "event_received" {
  source      = "../lambda"
  lambda_name = "event-received"
  description = "Function to react when an event is recieved"
  environment_variables = {
    ENVIRONMENT = data.aws_default_tags.current.tags.environment-name
    LPAS_TABLE  = var.lpas_table_name
  }
  image_uri   = "${var.event_received.lambda_function_image_ecr_url}:${var.event_received.lambda_function_image_tag}"
  ecr_arn     = var.lambda_function_image_ecr_arn
  environment = data.aws_default_tags.current.tags.environment-name
  kms_key     = data.aws_kms_alias.cloudwatch_application_logs_encryption.target_key_arn
  timeout     = 300
  memory      = 1024
  providers = {
    aws.region = aws.region
  }
}

resource "aws_iam_role_policy" "event_received" {
  name     = "event_received-${data.aws_default_tags.current.tags.environment-name}"
  role     = module.event_received.lambda_role.id
  policy   = data.aws_iam_policy_document.event_received.json
  provider = aws.region
}

data "aws_kms_alias" "dynamodb_encryption_key" {
  name     = "alias/${data.aws_default_tags.current.tags.application}_dynamodb_encryption"
  provider = aws.region
}

locals {
  policy_region_prefix = lower(replace(data.aws_region.current.name, "-", ""))
}

data "aws_iam_policy_document" "event_received" {
  statement {
    sid    = "${local.policy_region_prefix}DynamoDBEncryptionAccess"
    effect = "Allow"

    actions = [
      "kms:Encrypt",
      "kms:Decrypt",
      "kms:GenerateDataKey",
    ]

    resources = [
      data.aws_kms_alias.dynamodb_encryption_key.target_key_arn,
    ]
  }

  statement {
    sid = "${local.policy_region_prefix}Allow"

    actions = [
      "dynamodb:PutItem",
    ]

    resources = [
      var.lpas_table.arn,
      "${var.lpas_table.arn}/index/*",
    ]
  }

  provider = aws.region
}
