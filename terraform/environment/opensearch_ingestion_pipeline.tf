locals {
  enable_opensearch_ingestion_pipeline = true
}

data "aws_kms_alias" "dynamodb_encryption_key" {
  name     = "alias/${local.default_tags.application}_dynamodb_encryption"
  provider = aws.eu_west_1
}

data "aws_kms_alias" "opensearch_encryption_key" {
  name     = "alias/${local.default_tags.application}-opensearch-encryption-key"
  provider = aws.eu_west_1
}

data "aws_kms_alias" "dynamodb_exports_s3_bucket_encryption_key" {
  name     = "alias/${local.default_tags.application}-dynamodb-exports-s3-bucket-encryption"
  provider = aws.eu_west_1
}

data "aws_s3_bucket" "dynamodb_exports_bucket" {
  bucket   = "dynamodb-exports-${local.default_tags.application}-${local.default_tags.account-name}-eu-west-1"
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
    sid    = "CollectionActions"
    effect = "Allow"
    actions = [
      "aoss:BatchGetCollection",
      "aoss:APIAccessAll"
    ]
    resources = [
      data.aws_opensearchserverless_collection.lpas_collection.arn
    ]
  }

  statement {
    sid    = "WorkWithCollection"
    effect = "Allow"
    actions = [
      "aoss:GetSecurityPolicy",
      "aoss:CreateSecurityPolicy",
      "aoss:UpdateSecurityPolicy",
    ]
    resources = ["*"]
    condition {
      test     = "StringEquals"
      variable = "aoss:collection"
      values   = [data.aws_opensearchserverless_collection.lpas_collection.name]
    }
  }

  statement {
    sid    = "allowRunExportJob"
    effect = "Allow"
    actions = [
      "dynamodb:DescribeTable",
      "dynamodb:DescribeContinuousBackups",
      "dynamodb:ExportTableToPointInTime",
      "dynamodb:ListExports",
    ]
    resources = [
      aws_dynamodb_table.lpas_table.arn,
    ]
  }

  statement {
    sid    = "DescribeExports"
    effect = "Allow"
    actions = [
      "dynamodb:DescribeExport",
    ]
    resources = [
      "${aws_dynamodb_table.lpas_table.arn}/export/*",
    ]
  }

  statement {
    sid    = "DynamoDBAndExportEncryptionAccess"
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
      data.aws_kms_alias.dynamodb_exports_s3_bucket_encryption_key.target_key_arn
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
      "${aws_dynamodb_table.lpas_table.arn}/stream/*",
    ]
  }

  statement {
    sid    = "allowReadAndWriteToS3ForExport"
    effect = "Allow"
    actions = [
      "s3:HeadBucket",
      "s3:GetObject",
      "s3:CreateMultipartUpload",
      "s3:AbortMultipartUpload",
      "s3:UploadPart",
      "s3:PutObject",
      "s3:PutObjectAcl"
    ]
    resources = [
      data.aws_s3_bucket.dynamodb_exports_bucket.arn,
      "${data.aws_s3_bucket.dynamodb_exports_bucket.arn}/*",
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
  count       = local.enable_opensearch_ingestion_pipeline ? 1 : 0
  name_prefix = "${local.default_tags.environment-name}-opensearch-ingestion"
  description = "Security group for the opensearch ingestion pipeline"
  vpc_id      = data.aws_vpc.main.id
  provider    = aws.eu_west_1
}

# tfsec:ignore:aws-cloudwatch-log-group-customer-key
resource "aws_cloudwatch_log_group" "opensearch_pipeline" {
  count             = local.enable_opensearch_ingestion_pipeline ? 1 : 0
  name              = "/aws/vendedlogs/OpenSearchIngestion/lpas-${local.default_tags.environment-name}/audit-logs"
  retention_in_days = 1
  provider          = aws.eu_west_1
}

locals {
  data_protect_policy = file("cloudwatch_log_data_protection_policy/cloudwatch_log_data_protection_policy.json")
}
resource "aws_cloudwatch_log_data_protection_policy" "opensearch_pipeline" {
  log_group_name = aws_cloudwatch_log_group.opensearch_pipeline[0].name
  policy_document = jsonencode(merge(
    jsondecode(local.data_protect_policy),
    {
      Name = "data-protection-${local.default_tags.environment-name}-opensearch-ingestion"
    }
  ))
  provider = aws.eu_west_1
}

resource "aws_cloudwatch_query_definition" "opensearch_pipeline" {
  count           = local.enable_opensearch_ingestion_pipeline ? 1 : 0
  name            = "${local.default_tags.environment-name}/lpas-opensearch-pipeline"
  query_string    = "parse @message '* [*] * * - *' as timestamp, thread, Loglevel, endpoint, message | sort @timestamp desc | limit 1000"
  log_group_names = [aws_cloudwatch_log_group.opensearch_pipeline[0].name]
  provider        = aws.eu_west_1
}

data "aws_opensearchserverless_security_policy" "lpas_collection_network_policy" {
  name     = "policy-shared-${local.environment.account_name}"
  type     = "network"
  provider = aws.eu_west_1
}

locals {
  lpas_stream_pipeline_configuration_template_vars = {
    source = {
      tables = {
        table_arn         = aws_dynamodb_table.lpas_table.arn
        s3_bucket_name    = data.aws_s3_bucket.dynamodb_exports_bucket.id
        s3_sse_kms_key_id = data.aws_kms_alias.dynamodb_exports_s3_bucket_encryption_key.target_key_arn
        stream = {
          start_position = "LATEST"
        }
        aws = {
          sts_role_arn = module.global.iam_roles.opensearch_pipeline.arn
          region       = "eu-west-1"
        }
      }
    }
    routes = {
      lay_journey_lpas       = "'contains(/SK, \"DONOR#\") and contains(/PK, \"LPA#\")'"
      supporter_journey_lpas = "'contains(/SK, \"ORGANISATION#\") and contains(/PK, \"LPA#\")'"
    }

    sink = {
      opensearch = {
        hosts       = data.aws_opensearchserverless_collection.lpas_collection.collection_endpoint
        index       = "lpas_v2_${local.environment_name}"
        document_id = "$${/DocumentID}"
        aws = {
          sts_role_arn = module.global.iam_roles.opensearch_pipeline.arn
          region       = "eu-west-1"
          serverless_options = {
            network_policy_name = data.aws_opensearchserverless_security_policy.lpas_collection_network_policy.name
          }
        }
      }
    }
  }
}

resource "aws_opensearchserverless_access_policy" "pipeline" {
  count       = local.enable_opensearch_ingestion_pipeline ? 1 : 0
  name        = "pipeline-${local.environment_name}"
  type        = "data"
  description = "allow index and collection access for the opensearch ingestion pipeline"
  policy = jsonencode([
    {
      Rules = [
        {
          ResourceType = "index",
          Resource = [
            "index/shared-collection-${local.environment.account_name}/lpas_v2_${local.environment_name}",
          ],
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
  count                       = local.enable_opensearch_ingestion_pipeline ? 1 : 0
  pipeline_name               = "lpas-${local.default_tags.environment-name}-stream"
  max_units                   = 1
  min_units                   = 1
  pipeline_configuration_body = templatefile("opensearch_pipeline/lpas_stream_pipeline_configuration.yaml.tftpl", local.lpas_stream_pipeline_configuration_template_vars)
  buffer_options {
    persistent_buffer_enabled = false
  }
  log_publishing_options {
    cloudwatch_log_destination {
      log_group = aws_cloudwatch_log_group.opensearch_pipeline[0].name
    }
    is_logging_enabled = true
  }
  vpc_options {
    security_group_ids = [aws_security_group.opensearch_ingestion[0].id]
    subnet_ids         = data.aws_subnet.application[*].id
  }
  depends_on = [
    aws_opensearchserverless_access_policy.pipeline,
    aws_iam_role_policy.opensearch_pipeline,
    aws_security_group.opensearch_ingestion,
  ]

  provider = aws.eu_west_1
}
