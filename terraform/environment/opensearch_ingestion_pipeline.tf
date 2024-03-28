data "aws_kms_alias" "dynamodb_encryption_key" {
  name     = "alias/${local.default_tags.application}_dynamodb_encryption"
  provider = aws.eu_west_1
}

data "aws_kms_alias" "opensearch_encryption_key" {
  name     = "alias/${local.default_tags.application}-opensearch-encryption-key"
  provider = aws.eu_west_1
}

resource "aws_iam_role_policy" "opensearch_pipeline" {
  name     = "opensearch_pipeline"
  role     = module.global.iam_roles.opensearch_pipeline.name
  policy   = data.aws_iam_policy_document.opensearch_pipeline.json
  provider = aws.eu_west_1
}

data "aws_iam_policy_document" "opensearch_pipeline" {
  version = "2012-10-17"

  statement {
    sid    = "GetCollection"
    effect = "Allow"
    actions = [
      "aoss:BatchGetCollection",
    ]
    resources = ["*"]
  }

  statement {
    sid    = "WorkWithCollection"
    effect = "Allow"
    actions = [
      "aoss:CreateSecurityPolicy",
      "aoss:GetSecurityPolicy",
      "aoss:UpdateSecurityPolicy",
      "aoss:APIAccessAll"
    ]
    resources = ["*"]
    # condition {
    #   test     = "StringEquals"
    #   variable = "aoss:collection"
    #   values   = [aws_opensearchserverless_collection.lpas_collection.name]
    # }
    # condition {
    #   test     = "StringEquals"
    #   variable = "aws:SourceAccount"
    #   values   = [data.aws_caller_identity.eu_west_1.account_id]
    # }
    # condition {
    #   test     = "ArnLike"
    #   variable = "aws:SourceArn"
    #   values   = ["arn:aws:osis:eu-west-1:${data.aws_caller_identity.eu_west_1.account_id}:pipeline/*"]
    # }
  }

  statement {
    sid    = "allowRunExportJob"
    effect = "Allow"
    actions = [
      "dynamodb:DescribeTable",
      "dynamodb:DescribeContinuousBackups",
      "dynamodb:ExportTableToPointInTime",
    ]
    resources = [
      aws_dynamodb_table.lpas_table.arn,
    ]
  }

  statement {
    sid    = "DynamoDBEncryptionAccess"
    effect = "Allow"
    actions = [
      "kms:Decrypt",
      "kms:GenerateDataKey",
    ]
    resources = [
      data.aws_kms_alias.dynamodb_encryption_key.target_key_arn,
    ]
  }

  statement {
    sid    = "OpensearchEncryptionAccess"
    effect = "Allow"
    actions = [
      "kms:Encrypt",
      "kms:Decrypt",
      "kms:GenerateDataKey",
    ]
    resources = [
      data.aws_kms_alias.opensearch_encryption_key.target_key_arn,
    ]
  }


  statement {
    sid    = "allowReadFromStream"
    effect = "Allow"
    actions = [
      "dynamodb:DescribeStream",
      "dynamodb:GetRecords",
      "dynamodb:GetShardIterator",
    ]
    resources = [
      # "*",
      "${aws_dynamodb_table.lpas_table.arn}/stream/*",
    ]
  }
}

data "aws_vpc" "main" {
  filter {
    name   = "tag:application"
    values = [local.default_tags.application]
  }
  provider = aws.eu_west_1
}

data "aws_availability_zones" "available" {
  provider = aws.eu_west_1
}

data "aws_subnet" "application" {
  count             = 3
  vpc_id            = data.aws_vpc.main.id
  availability_zone = data.aws_availability_zones.available.names[count.index]

  filter {
    name   = "tag:Name"
    values = ["application*"]
  }
  provider = aws.eu_west_1
}

resource "aws_security_group" "opensearch_ingestion" {
  name_prefix = "${local.default_tags.environment-name}-opensearch-ingestion"
  description = "Security group for the opensearch ingestion pipeline"
  vpc_id      = data.aws_vpc.main.id
  provider    = aws.eu_west_1
}

resource "aws_cloudwatch_log_group" "opensearch_pipeline" {
  name              = "/aws/vendedlogs/OpenSearchIngestion/lpas-${local.default_tags.environment-name}/audit-logs"
  retention_in_days = 1
  provider          = aws.eu_west_1
}

resource "aws_cloudwatch_query_definition" "opensearch_pipeline" {
  name            = "${local.default_tags.environment-name}/lpas-opensearch-pipeline"
  query_string    = "parse @message '* [*] * * - *' as timestamp, thread, Loglevel, endpoint, message | sort @timestamp desc | limit 1000"
  log_group_names = [aws_cloudwatch_log_group.opensearch_pipeline.name]
  provider        = aws.eu_west_1
}

locals {
  pipeline_configuration_template_vars = {
    source = {
      tables = {
        table_arn = aws_dynamodb_table.lpas_table.arn
        stream = {
          start_position = "LATEST"
        }
        aws = {
          sts_role_arn = module.global.iam_roles.opensearch_pipeline.arn
          region       = "eu-west-1"
        }
      }
    }
    supporter_lpas = {
      route = "'contains(/\"SK\", \"ORGANISATION#\") and contains(/\"PK\", \"LPA#\")'"
      sink = {
        opensearch = {
          hosts = aws_opensearchserverless_collection.lpas_collection.collection_endpoint
          index = "lpas"
          aws = {
            sts_role_arn = module.global.iam_roles.opensearch_pipeline.arn
            region       = "eu-west-1"
            serverless_options = {
              network_policy_name = aws_opensearchserverless_security_policy.lpas_collection_network_policy.name
            }
          }
        }
      }
    }
  }
}

resource "aws_opensearchserverless_access_policy" "pipeline" {
  name        = "pipeline-${local.environment_name}"
  type        = "data"
  description = "allow index and collection access for the opensearch ingestion pipeline"
  policy = jsonencode([
    {
      Rules = [
        {
          ResourceType = "index",
          Resource     = ["index/collection-${local.environment_name}/*"],
          Permission = [
            "aoss:CreateIndex",
            "aoss:UpdateIndex",
            "aoss:DescribeIndex",
            "aoss:WriteDocument",
          ]
        }
      ],
      Principal = [
        module.global.iam_roles.opensearch_pipeline.arn
      ]
    }
  ])
  provider = aws.eu_west_1
}

resource "aws_osis_pipeline" "lpas_stream" {
  pipeline_name               = "lpas-${local.default_tags.environment-name}"
  max_units                   = 1
  min_units                   = 1
  pipeline_configuration_body = templatefile("opensearch_pipeline/pipeline_configuration.yaml.tftpl", local.pipeline_configuration_template_vars)
  buffer_options {
    persistent_buffer_enabled = false
  }
  log_publishing_options {
    cloudwatch_log_destination {
      log_group = aws_cloudwatch_log_group.opensearch_pipeline.name
    }
    is_logging_enabled = true
  }
  vpc_options {
    security_group_ids = [aws_security_group.opensearch_ingestion.id]
    subnet_ids         = data.aws_subnet.application[*].id
  }
  provider = aws.eu_west_1
}

moved {
  from = aws_osis_pipeline.lpas_stream[0]
  to   = aws_osis_pipeline.lpas_stream
}
moved {
  from = aws_cloudwatch_log_group.opensearch_pipeline[0]
  to   = aws_cloudwatch_log_group.opensearch_pipeline
}
