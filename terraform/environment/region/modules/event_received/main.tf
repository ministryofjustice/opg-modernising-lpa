module "event_received" {
  source      = "../lambda"
  lambda_name = "event-received"
  description = "Function to react when an event is received"
  environment_variables = {
    LPAS_TABLE                 = var.lpas_table.name
    GOVUK_NOTIFY_IS_PRODUCTION = data.aws_default_tags.current.tags.environment-name == "production" ? "1" : "0"
    GOVUK_NOTIFY_BASE_URL      = "https://api.notifications.service.gov.uk"
    APP_PUBLIC_URL             = "https://${var.app_public_url}"
    UPLOADS_S3_BUCKET_NAME     = var.uploads_bucket.bucket
    UID_BASE_URL               = var.uid_base_url
    LPA_STORE_BASE_URL         = var.lpa_store_base_url
    SEARCH_ENDPOINT            = var.search_endpoint
    SEARCH_INDEX_NAME          = var.search_index_name
    SEARCH_INDEXING_DISABLED   = 1
    EVENT_BUS_NAME             = var.event_bus_name
    JWT_KEY_SECRET_ARN         = data.aws_secretsmanager_secret.lpa_store_jwt_key.arn
  }
  image_uri            = "${var.lambda_function_image_ecr_url}:${var.lambda_function_image_tag}"
  aws_iam_role         = var.event_received_lambda_role
  environment          = data.aws_default_tags.current.tags.environment-name
  kms_key              = data.aws_kms_alias.cloudwatch_application_logs_encryption.target_key_arn
  iam_policy_documents = [data.aws_iam_policy_document.api_access_policy.json]
  timeout              = 300
  memory               = 1024
  vpc_config = {
    subnet_ids         = var.vpc_config.subnet_ids
    security_group_ids = var.vpc_config.security_group_ids
  }
  providers = {
    aws.region = aws.region
  }
}

# data "aws_kms_alias" "sqs" {
#   name     = "alias/${data.aws_default_tags.current.tags.application}_sqs_secret_encryption_key"
#   provider = aws.region
# }

resource "aws_sqs_queue" "receive_events_queue" {
  name                              = "${data.aws_default_tags.current.tags.environment-name}-receive-events-queue"
  # kms_master_key_id                 = data.aws_kms_alias.sqs.target_key_id
  # kms_data_key_reuse_period_seconds = 300

  visibility_timeout_seconds = 300
  redrive_policy = jsonencode({
    deadLetterTargetArn = aws_sqs_queue.receive_events_deadletter.arn
    maxReceiveCount     = 3
  })
  policy   = data.aws_iam_policy_document.receive_events_queue_policy.json
  provider = aws.region
}

data "aws_iam_policy_document" "receive_events_queue_policy" {
  statement {
    sid    = "${local.policy_region_prefix}Send"
    effect = "Allow"

    principals {
      type        = "Service"
      identifiers = ["events.amazonaws.com"]
    }

    actions = ["sqs:SendMessage"]
    # resources = [aws_sqs_queue.receive_events_queue.arn]
    resources = ["*"]

    # condition {
    #   test     = "ArnEquals"
    #   variable = "aws:SourceArn"
    #   values   = [
    #     aws_cloudwatch_event_rule.receive_events_sirius,
    #     aws_cloudwatch_event_rule.receive_events_lpa_store,
    #     aws_cloudwatch_event_rule.receive_events_mlpa,
    #   ]
    # }
  }
}

resource "aws_sqs_queue" "receive_events_deadletter" {
  name                              = "${data.aws_default_tags.current.tags.environment-name}-receive-events-deadletter"
  # kms_master_key_id                 = data.aws_kms_alias.sqs.target_key_id
  # kms_data_key_reuse_period_seconds = 300
  provider                          = aws.region
}

resource "aws_sqs_queue_redrive_allow_policy" "receive_events_redrive_allow_policy" {
  queue_url = aws_sqs_queue.receive_events_deadletter.id

  redrive_allow_policy = jsonencode({
    redrivePermission = "byQueue",
    sourceQueueArns   = [aws_sqs_queue.receive_events_queue.arn]
  })
  provider = aws.region
}

resource "aws_lambda_event_source_mapping" "reveive_events_mapping" {
  event_source_arn = aws_sqs_queue.receive_events_queue.arn
  enabled          = true
  function_name    = module.event_received.lambda.arn
  provider         = aws.region
}

data "aws_iam_policy_document" "api_access_policy" {
  statement {
    sid       = "allowApiAccess"
    effect    = "Allow"
    resources = var.allowed_api_arns
    actions = [
      "execute-api:Invoke",
    ]
  }
}

resource "aws_cloudwatch_event_rule" "receive_events_sirius" {
  name           = "${data.aws_default_tags.current.tags.environment-name}-receive-events-sirius"
  description    = "receive events from sirius"
  event_bus_name = var.event_bus_name

  event_pattern = jsonencode({
    source = ["opg.poas.sirius"],
    detail-type = [
      "certificate-provider-submission-completed",
      "donor-submission-completed",
      "evidence-received",
      "further-info-requested",
      "reduced-fee-approved",
      "reduced-fee-declined",
    ],
  })
  provider = aws.region
}

resource "aws_cloudwatch_event_target" "receive_events_sirius" {
  target_id      = "${data.aws_default_tags.current.tags.environment-name}-receive-events-sirius"
  event_bus_name = var.event_bus_name
  rule           = aws_cloudwatch_event_rule.receive_events_sirius.name
  arn            = aws_sqs_queue.receive_events_queue.arn
  provider       = aws.region
  dead_letter_config {
    arn = var.event_bus_dead_letter_queue.arn
  }
}

resource "aws_cloudwatch_event_rule" "receive_events_lpa_store" {
  name           = "${data.aws_default_tags.current.tags.environment-name}-receive-events-lpa-store"
  description    = "receive events from lpa store"
  event_bus_name = var.event_bus_name

  event_pattern = jsonencode({
    source      = ["opg.poas.lpastore"],
    detail-type = ["lpa-updated"],
  })
  provider = aws.region
}

resource "aws_cloudwatch_event_target" "receive_events_lpa_store" {
  target_id      = "${data.aws_default_tags.current.tags.environment-name}-receive-events-lpa-store"
  event_bus_name = var.event_bus_name
  rule           = aws_cloudwatch_event_rule.receive_events_lpa_store.name
  arn            = aws_sqs_queue.receive_events_queue.arn
  dead_letter_config {
    arn = var.event_bus_dead_letter_queue.arn
  }
  provider = aws.region
}

resource "aws_cloudwatch_event_rule" "receive_events_mlpa" {
  name           = "${data.aws_default_tags.current.tags.environment-name}-receive-events-mlpa"
  description    = "receive events from mlpa"
  event_bus_name = var.event_bus_name

  event_pattern = jsonencode({
    source      = ["opg.poas.makeregister"],
    detail-type = ["uid-requested"],
  })
  provider = aws.region
}

resource "aws_cloudwatch_event_target" "receive_events_mlpa" {
  target_id      = "${data.aws_default_tags.current.tags.environment-name}-receive-events-mlpa"
  event_bus_name = var.event_bus_name
  rule           = aws_cloudwatch_event_rule.receive_events_mlpa.name
  arn            = aws_sqs_queue.receive_events_queue.arn
  dead_letter_config {
    arn = var.event_bus_dead_letter_queue.arn
  }
  provider = aws.region
}

resource "aws_lambda_permission" "allow_cloudwatch_to_call_event_received_sirius" {
  statement_id   = "AllowExecutionFromCloudWatchSirius"
  action         = "lambda:InvokeFunction"
  function_name  = module.event_received.lambda.function_name
  principal      = "events.amazonaws.com"
  source_account = data.aws_caller_identity.current.account_id
  source_arn     = aws_cloudwatch_event_rule.receive_events_sirius.arn
  provider       = aws.region
}

resource "aws_lambda_permission" "allow_cloudwatch_to_call_event_received_mlpa" {
  statement_id   = "AllowExecutionFromCloudWatchMlpa"
  action         = "lambda:InvokeFunction"
  function_name  = module.event_received.lambda.function_name
  principal      = "events.amazonaws.com"
  source_account = data.aws_caller_identity.current.account_id
  source_arn     = aws_cloudwatch_event_rule.receive_events_mlpa.arn
  provider       = aws.region
}

resource "aws_iam_role_policy" "event_received" {
  name     = "event_received-${data.aws_default_tags.current.tags.environment-name}"
  role     = var.event_received_lambda_role.id
  policy   = data.aws_iam_policy_document.event_received.json
  provider = aws.region
}

resource "aws_iam_role_policy_attachment" "cloudwatch_lambda_insights" {
  role       = var.event_received_lambda_role.id
  policy_arn = "arn:aws:iam::aws:policy/CloudWatchLambdaInsightsExecutionRolePolicy"
  provider   = aws.region
}

data "aws_kms_alias" "dynamodb_encryption_key" {
  name     = "alias/${data.aws_default_tags.current.tags.application}_dynamodb_encryption"
  provider = aws.region
}

data "aws_kms_alias" "secrets_manager_secret_encryption_key" {
  name     = "alias/${data.aws_default_tags.current.tags.application}_secrets_manager_secret_encryption_key"
  provider = aws.region
}

data "aws_kms_alias" "aws_lambda" {
  name     = "alias/aws/lambda"
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
      "kms:RetireGrant",
    ]

    resources = [
      data.aws_kms_alias.dynamodb_encryption_key.target_key_arn,
    ]
  }

  statement {
    sid = "${local.policy_region_prefix}Allow"

    actions = [
      "dynamodb:PutItem",
      "dynamodb:Query",
      "dynamodb:GetItem",
      "dynamodb:UpdateItem",
    ]

    resources = [
      var.lpas_table.arn,
      "${var.lpas_table.arn}/index/*",
    ]
  }

  statement {
    sid    = "${local.policy_region_prefix}SecretAccess"
    effect = "Allow"

    actions = [
      "secretsmanager:GetSecretValue",
      "secretsmanager:DescribeSecret",
    ]

    resources = [
      data.aws_secretsmanager_secret.gov_uk_notify_api_key.arn,
      data.aws_secretsmanager_secret.lpa_store_jwt_secret_key.arn,
      data.aws_secretsmanager_secret.lpa_store_jwt_key.arn,
    ]
  }

  statement {
    effect = "Allow"

    resources = [
      data.aws_kms_alias.secrets_manager_secret_encryption_key.target_key_arn,
      data.aws_kms_alias.aws_lambda.target_key_arn,
      data.aws_kms_alias.jwt_key.target_key_arn,
      # data.aws_kms_alias.sqs.target_key_arn,
    ]

    actions = [
      "kms:Decrypt",
      "kms:GenerateDataKey",
      "kms:GenerateDataKeyPair",
      "kms:GenerateDataKeyPairWithoutPlaintext",
      "kms:GenerateDataKeyWithoutPlaintext",
      "kms:DescribeKey",
    ]
  }

  statement {
    effect = "Allow"

    resources = [
      "${var.uploads_bucket.arn}/*"
    ]

    actions = [
      "s3:getObjectTagging",
    ]
  }

  statement {
    sid    = "${local.policy_region_prefix}OpenSearchAccess"
    effect = "Allow"

    actions = [
      "aoss:APIAccessAll"
    ]

    resources = [
      var.search_collection_arn
    ]
  }

  statement {
    sid    = "${local.policy_region_prefix}CrossAccountPutAccess"
    effect = "Allow"
    actions = [
      "events:PutEvents",
    ]
    resources = [
      var.event_bus_arn
    ]
  }

  statement {
    sid       = "${local.policy_region_prefix}SqsAccess"
    effect    = "Allow"
    actions   = ["sqs:ReceiveMessage", "sqs:DeleteMessage", "sqs:GetQueueAttributes"]
    resources = [aws_sqs_queue.receive_events_queue.arn]
  }

  provider = aws.region
}
