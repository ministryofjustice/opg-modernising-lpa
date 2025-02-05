resource "aws_opensearchserverless_security_policy" "lpas_collection_encryption_policy" {
  name        = "policy-shared-${data.aws_default_tags.current.tags.account-name}"
  type        = "encryption"
  description = "encryption policy for collection"
  policy = jsonencode({
    Rules = [
      {
        Resource     = ["collection/shared-collection-${data.aws_default_tags.current.tags.account-name}"],
        ResourceType = "collection"
      }
    ],
    AWSOwnedKey = false
    KmsARN      = var.opensearch_kms_target_key_arn
  })
  provider = aws.region
}

resource "aws_opensearchserverless_collection" "lpas_collection" {
  name       = "shared-collection-${data.aws_default_tags.current.tags.account-name}"
  type       = "SEARCH"
  depends_on = [aws_opensearchserverless_security_policy.lpas_collection_encryption_policy]
  provider   = aws.region
}

resource "aws_opensearchserverless_security_policy" "lpas_collection_network_policy" {
  name        = "policy-shared-${data.aws_default_tags.current.tags.account-name}"
  type        = "network"
  description = "VPC access for collection endpoint"
  policy = jsonencode([
    {
      Description = "VPC access for collection endpoint",
      Rules = [
        {
          ResourceType = "collection",
          Resource     = ["collection/shared-collection-${data.aws_default_tags.current.tags.account-name}"]
        }
      ],
      AllowFromPublic = false,
      SourceVPCEs = [
        aws_opensearchserverless_vpc_endpoint.lpas_collection_vpc_endpoint.id,
      ]
    },
    {
      AllowFromPublic = true
      Description     = "public access to dashboard"
      Rules = [
        {
          Resource     = ["collection/shared-collection-${data.aws_default_tags.current.tags.account-name}"]
          ResourceType = "dashboard"
        }
      ]
    }
  ])
  provider = aws.region
}

resource "aws_opensearchserverless_security_policy" "lpas_collection_development_network_policy" {
  count       = data.aws_default_tags.current.tags.account-name == "development" ? 1 : 0
  name        = "development-public-access"
  type        = "network"
  description = "Public access for development collection endpoints"
  policy = jsonencode([
    {
      Description = "Public access for development collection endpoint",
      Rules = [
        {
          ResourceType = "collection",
          Resource     = ["collection/shared-collection-development"]
        }
      ],
      AllowFromPublic = true,
    },
  ])
  provider = aws.region
}

resource "aws_opensearchserverless_access_policy" "github_actions_access" {
  count       = data.aws_default_tags.current.tags.account-name == "development" ? 1 : 0
  name        = "github-access-shared-development"
  type        = "data"
  description = "allow index and collection access for team"
  policy = jsonencode([
    {
      Rules = [
        {
          ResourceType = "index",
          Resource     = ["index/shared-collection-development/*"],
          Permission   = ["aoss:*"]
        },
        {
          ResourceType = "collection",
          Resource     = ["collection/shared-collection-development"],
          Permission   = ["aoss:*"]
        }
      ],
      Principal = [
        "arn:aws:iam::${data.aws_caller_identity.current.account_id}:role/modernising-lpa-github-actions-opensearch-delete-index"
      ]
    }
  ])
  provider = aws.region
}

resource "aws_opensearchserverless_access_policy" "team_operator_access" {
  count       = data.aws_default_tags.current.tags.account-name == "production" ? 0 : 1
  name        = "team-access-shared-${data.aws_default_tags.current.tags.account-name}"
  type        = "data"
  description = "allow index and collection access for team"
  policy = jsonencode([
    {
      Rules = [
        {
          ResourceType = "index",
          Resource     = ["index/shared-collection-${data.aws_default_tags.current.tags.account-name}/*"],
          Permission   = ["aoss:*"]
        },
        {
          ResourceType = "collection",
          Resource     = ["collection/shared-collection-${data.aws_default_tags.current.tags.account-name}"],
          Permission   = ["aoss:*"]
        }
      ],
      Principal = [
        "arn:aws:iam::${data.aws_caller_identity.current.account_id}:role/operator"
      ]
    }
  ])
  provider = aws.region
}

resource "aws_opensearchserverless_access_policy" "team_breakglass_access" {
  count       = data.aws_default_tags.current.tags.account-name == "production" ? 1 : 0
  name        = "team-access-shared-${data.aws_default_tags.current.tags.account-name}"
  type        = "data"
  description = "allow index and collection access for team"
  policy = jsonencode([
    {
      Rules = [
        {
          ResourceType = "index",
          Resource     = ["index/shared-collection-${data.aws_default_tags.current.tags.account-name}/*"],
          Permission   = ["aoss:*"]
        },
        {
          ResourceType = "collection",
          Resource     = ["collection/shared-collection-${data.aws_default_tags.current.tags.account-name}"],
          Permission   = ["aoss:*"]
        }
      ],
      Principal = [
        "arn:aws:iam::${data.aws_caller_identity.current.account_id}:role/breakglass"
      ]
    }
  ])
  provider = aws.region
}

resource "aws_cloudwatch_metric_alarm" "opensearch_4xx_errors" {
  alarm_name                = "${data.aws_default_tags.current.tags.account-name}-opensearch-4xx-errors"
  alarm_actions             = [aws_sns_topic.opensearch.arn]
  comparison_operator       = "GreaterThanOrEqualToThreshold"
  evaluation_periods        = "1"
  metric_name               = "4xx"
  namespace                 = "AWS/AOSS"
  period                    = "30"
  statistic                 = "Maximum"
  threshold                 = "1"
  alarm_description         = "This metric monitors AWS OpenSearch Service 4xx error count for ${data.aws_default_tags.current.tags.account-name}"
  insufficient_data_actions = []
  dimensions = {
    CollectionId   = aws_opensearchserverless_collection.lpas_collection.id
    CollectionName = aws_opensearchserverless_collection.lpas_collection.name
  }
  provider = aws.region
}

resource "aws_cloudwatch_metric_alarm" "opensearch_5xx_errors" {
  alarm_name                = "${data.aws_default_tags.current.tags.account-name}-opensearch-5xx-errors"
  alarm_actions             = [aws_sns_topic.opensearch.arn]
  comparison_operator       = "GreaterThanOrEqualToThreshold"
  evaluation_periods        = "1"
  metric_name               = "5xx"
  namespace                 = "AWS/AOSS"
  period                    = "30"
  statistic                 = "Maximum"
  threshold                 = "1"
  alarm_description         = "This metric monitors AWS OpenSearch Service 5xx error count for ${data.aws_default_tags.current.tags.account-name}"
  insufficient_data_actions = []
  dimensions = {
    CollectionId   = aws_opensearchserverless_collection.lpas_collection.id
    CollectionName = aws_opensearchserverless_collection.lpas_collection.name
  }
  provider = aws.region
}

data "pagerduty_vendor" "cloudwatch" {
  name = "Cloudwatch"
}

data "pagerduty_service" "main" {
  name = var.pagerduty_service_name
}

resource "aws_sns_topic" "opensearch" {
  name                                     = "${data.aws_default_tags.current.tags.account-name}-opensearch-alarms"
  kms_master_key_id                        = var.sns_kms_key.eu_west_1_target_key_id
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
  provider                                 = aws.region
}

resource "pagerduty_service_integration" "opensearch" {
  name    = "Modernising LPA Shared ${data.aws_default_tags.current.tags.account-name} OpenSearch Alarm"
  service = data.pagerduty_service.main.id
  vendor  = data.pagerduty_vendor.cloudwatch.id
}

resource "aws_sns_topic_subscription" "opensearch" {
  topic_arn              = aws_sns_topic.opensearch.arn
  protocol               = "https"
  endpoint_auto_confirms = true
  endpoint               = "https://events.pagerduty.com/integration/${pagerduty_service_integration.opensearch.integration_key}/enqueue"
  provider               = aws.region
}
