data "aws_vpc_endpoint" "opensearch" {
  tags = {
    Name = "opensearch-eu-west-1"
  }
  provider = aws.eu_west_1
}

data "aws_kms_alias" "opensearch" {
  name     = "alias/${local.default_tags.application}-opensearch-encryption-key"
  provider = aws.eu_west_1
}

resource "aws_opensearchserverless_security_policy" "lpas_collection_encryption_policy" {
  name        = "policy-${local.environment_name}"
  type        = "encryption"
  description = "encryption policy for collection"
  policy = jsonencode({
    Rules = [
      {
        Resource     = ["collection/collection-${local.environment_name}"],
        ResourceType = "collection"
      }
    ],
    AWSOwnedKey = false
    KmsARN      = data.aws_kms_alias.opensearch.target_key_arn
  })
  provider = aws.eu_west_1
}

resource "aws_opensearchserverless_collection" "lpas_collection" {
  name       = "collection-${local.environment_name}"
  type       = "SEARCH"
  depends_on = [aws_opensearchserverless_security_policy.lpas_collection_encryption_policy]
  provider   = aws.eu_west_1
}

resource "aws_opensearchserverless_security_policy" "lpas_collection_network_policy" {
  name        = "policy-${local.environment_name}"
  type        = "network"
  description = "VPC access for collection endpoint"
  policy = jsonencode([
    {
      Description = "VPC access for collection endpoint",
      Rules = [
        {
          ResourceType = "collection",
          Resource     = ["collection/collection-${local.environment_name}"]
        }
      ],
      AllowFromPublic = false,
      SourceVPCEs = [
        data.aws_vpc_endpoint.opensearch.id
      ]
    },
    {
      AllowFromPublic = true
      Description     = "public access to dashboard"
      Rules = [
        {
          Resource     = ["collection/collection-${local.environment_name}"]
          ResourceType = "dashboard"
        }
      ]
    }
  ])
  provider = aws.eu_west_1
}

resource "aws_opensearchserverless_access_policy" "app" {
  name        = "app-${local.environment_name}"
  type        = "data"
  description = "allow index and collection access"
  policy = jsonencode([
    {
      Rules = [
        {
          ResourceType = "index",
          Resource     = ["index/collection-${local.environment_name}/*"],
          Permission   = ["aoss:*"]
        },
        {
          ResourceType = "collection",
          Resource     = ["collection/collection-${local.environment_name}"],
          Permission   = ["aoss:*"]
        }
      ],
      Principal = [
        module.global.iam_roles.app_ecs_task_role.arn,
        "arn:aws:iam::${data.aws_caller_identity.eu_west_1.account_id}:role/operator"
      ]
    }
  ])
  provider = aws.eu_west_1
}

resource "aws_opensearchserverless_access_policy" "event_received" {
  name        = "event-received-${local.environment_name}"
  type        = "data"
  description = "allow index and collection access"
  policy = jsonencode([
    {
      Rules = [
        {
          ResourceType = "index",
          Resource     = ["index/collection-${local.environment_name}/*"],
          Permission   = ["aoss:*"]
        },
        {
          ResourceType = "collection",
          Resource     = ["collection/collection-${local.environment_name}"],
          Permission   = ["aoss:*"]
        }
      ],
      Principal = [module.global.iam_roles.event_received_lambda.arn]
    }
  ])
  provider = aws.eu_west_1
}

resource "aws_opensearchserverless_access_policy" "team_operator_access" {
  count       = local.environment_name == "production" ? 0 : 1
  name        = "team-access-${local.environment_name}"
  type        = "data"
  description = "allow index and collection access for team"
  policy = jsonencode([
    {
      Rules = [
        {
          ResourceType = "index",
          Resource     = ["index/collection-${local.environment_name}/*"],
          Permission   = ["aoss:*"]
        },
        {
          ResourceType = "collection",
          Resource     = ["collection/collection-${local.environment_name}"],
          Permission   = ["aoss:*"]
        }
      ],
      Principal = [
        "arn:aws:iam::${data.aws_caller_identity.eu_west_1.account_id}:role/operator"
      ]
    }
  ])
  provider = aws.eu_west_1
}

resource "aws_opensearchserverless_access_policy" "team_breakglas_access" {
  count       = local.environment_name == "production" ? 1 : 0
  name        = "team-access-${local.environment_name}"
  type        = "data"
  description = "allow index and collection access for team"
  policy = jsonencode([
    {
      Rules = [
        {
          ResourceType = "index",
          Resource     = ["index/collection-${local.environment_name}/*"],
          Permission   = ["aoss:*"]
        },
        {
          ResourceType = "collection",
          Resource     = ["collection/collection-${local.environment_name}"],
          Permission   = ["aoss:*"]
        }
      ],
      Principal = [
        "arn:aws:iam::${data.aws_caller_identity.eu_west_1.account_id}:role/breakglass"
      ]
    }
  ])
  provider = aws.eu_west_1
}

resource "aws_cloudwatch_metric_alarm" "opensearch_4xx_errors" {
  alarm_name                = "${local.environment_name}-opensearch-4xx-errors"
  alarm_actions             = [aws_sns_topic.opensearch.arn]
  comparison_operator       = "GreaterThanOrEqualToThreshold"
  evaluation_periods        = "1"
  metric_name               = "4xx"
  namespace                 = "AWS/AOSS"
  period                    = "30"
  statistic                 = "Maximum"
  threshold                 = "1"
  alarm_description         = "This metric monitors AWS OpenSearch Service 4xx error count for ${local.environment_name}"
  insufficient_data_actions = []
  dimensions = {
    CollectionId   = aws_opensearchserverless_collection.lpas_collection.id
    CollectionName = aws_opensearchserverless_collection.lpas_collection.name
  }
  provider = aws.eu_west_1
}

resource "aws_cloudwatch_metric_alarm" "opensearch_5xx_errors" {
  alarm_name                = "${local.environment_name}-opensearch-5xx-errors"
  alarm_actions             = [aws_sns_topic.opensearch.arn]
  comparison_operator       = "GreaterThanOrEqualToThreshold"
  evaluation_periods        = "1"
  metric_name               = "5xx"
  namespace                 = "AWS/AOSS"
  period                    = "30"
  statistic                 = "Maximum"
  threshold                 = "1"
  alarm_description         = "This metric monitors AWS OpenSearch Service 5xx error count for ${local.environment_name}"
  insufficient_data_actions = []
  dimensions = {
    CollectionId   = aws_opensearchserverless_collection.lpas_collection.id
    CollectionName = aws_opensearchserverless_collection.lpas_collection.name
  }
  provider = aws.eu_west_1
}

data "pagerduty_vendor" "cloudwatch" {
  name = "Cloudwatch"
}

data "pagerduty_service" "main" {
  name = local.environment.pagerduty_service_name
}

data "aws_kms_alias" "sns_kms_key_alias" {
  name     = "alias/${local.environment_name}_sns_secret_encryption_key"
  provider = aws.eu_west_1
}

resource "aws_sns_topic" "opensearch" {
  name                                     = "${local.environment_name}-opensearch-alarms"
  kms_master_key_id                        = data.aws_kms_alias.sns_kms_key_alias.target_key_id
  application_failure_feedback_role_arn    = data.aws_iam_role.sns_failure_feedback.arn
  application_success_feedback_role_arn    = data.aws_iam_role.sns_success_feedback.arn
  application_success_feedback_sample_rate = 100
  firehose_failure_feedback_role_arn       = data.aws_iam_role.sns_failure_feedback.arn
  firehose_success_feedback_role_arn       = data.aws_iam_role.sns_success_feedback.arn
  firehose_success_feedback_sample_rate    = 100
  http_failure_feedback_role_arn           = data.aws_iam_role.sns_failure_feedback.arn
  http_success_feedback_role_arn           = data.aws_iam_role.sns_success_feedback.arn
  http_success_feedback_sample_rate        = 100
  lambda_failure_feedback_role_arn         = data.aws_iam_role.sns_failure_feedback.arn
  lambda_success_feedback_role_arn         = data.aws_iam_role.sns_success_feedback.arn
  lambda_success_feedback_sample_rate      = 100
  sqs_failure_feedback_role_arn            = data.aws_iam_role.sns_failure_feedback.arn
  sqs_success_feedback_role_arn            = data.aws_iam_role.sns_success_feedback.arn
  sqs_success_feedback_sample_rate         = 100
  provider                                 = aws.eu_west_1
}

resource "pagerduty_service_integration" "opensearch" {
  name    = "Modernising LPA ${local.environment_name} OpenSearch Alarm"
  service = data.pagerduty_service.main.id
  vendor  = data.pagerduty_vendor.cloudwatch.id
}

resource "aws_sns_topic_subscription" "opensearch" {
  topic_arn              = aws_sns_topic.opensearch.arn
  protocol               = "https"
  endpoint_auto_confirms = true
  endpoint               = "https://events.pagerduty.com/integration/${pagerduty_service_integration.opensearch.integration_key}/enqueue"
  provider               = aws.eu_west_1
}
