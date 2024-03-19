resource "aws_iam_role_policy" "opensearch_ingestion" {
  name     = "opensearch_ingestion"
  role     = module.global.iam_roles.opensearch_ingestion.name
  policy   = data.aws_iam_policy_document.opensearch_ingestion.json
  provider = aws.global
}

data "aws_iam_policy_document" "opensearch_ingestion" {
  version = "2012-10-17"

  statement {
    sid    = "WorkWithCollection"
    effect = "Allow"
    actions = [
      "aoss:BatchGetCollection",
    ]
    resources = [
      aws_opensearchserverless_collection.lpas_collection.arn,
    ]
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
    sid    = "allowCheckExportjob"
    effect = "Allow"
    actions = [
      "dynamodb:DescribeExport",
    ]
    resources = [
      "${aws_dynamodb_table.lpas_table.arn}/export/*",
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
      aws_dynamodb_table.lpas_table.stream_arn,
    ]
  }

  statement {
    sid    = "allowReadAndWriteToS3ForExport"
    effect = "Allow"
    actions = [
      "s3:GetObject",
      "s3:AbortMultipartUpload",
      "s3:PutObject",
      "s3:PutObjectAcl",
    ]
    resources = [
      "${aws_s3_bucket.opensearch_ingestion.arn}/export/*",
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
  vpc_id      = data.aws_vpc.main.id
  provider    = aws.eu_west_1
}

resource "aws_cloudwatch_log_group" "opensearch_ingestion" {
  name              = "/aws/vendedlogs/OpenSearchIngestion/lpas-${local.default_tags.environment-name}/audit-logs"
  retention_in_days = 1
  provider          = aws.eu_west_1
}

resource "aws_s3_bucket" "opensearch_ingestion" {
  bucket   = "${local.default_tags.environment-name}-opensearch-ingestion"
  provider = aws.eu_west_1
}

locals {
  pipeline_configuration_tempalte_vars = {
    source = {
      tables = {
        table_arn = aws_dynamodb_table.lpas_table.arn
        stream = {
          start_position = "LATEST"
        }
        export = {
          s3_bucket = aws_s3_bucket.opensearch_ingestion.bucket
          s3_region = "eu-west-1"
          s3_prefix = "${local.default_tags.environment-name}/ddb-to-opensearch-export/"
        }
        aws = {
          sts_role_arn = module.global.iam_roles.opensearch_ingestion.arn
          region       = "eu-west-1"
        }
      }
    }
    supporter_lpas = {
      route = "'contains(/SK, \"ORGANISATION#\")'"
      sink = {
        opensearch = {
          hosts = aws_opensearchserverless_collection.lpas_collection.collection_endpoint
          index = "lpas2"
          aws = {
            sts_role_arn = module.global.iam_roles.opensearch_ingestion.arn
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

resource "aws_osis_pipeline" "example" {
  pipeline_name               = "lpas-${local.default_tags.environment-name}"
  max_units                   = 1
  min_units                   = 1
  pipeline_configuration_body = templatefile("opensearch_pipeline/pipeline_configuration.yaml.tftpl", local.pipeline_configuration_tempalte_vars)
  buffer_options {
    persistent_buffer_enabled = false
  }
  log_publishing_options {
    cloudwatch_log_destination {
      log_group = aws_cloudwatch_log_group.opensearch_ingestion.name
    }
    is_logging_enabled = true
  }
  vpc_options {
    security_group_ids = [aws_security_group.opensearch_ingestion.id]
    subnet_ids         = data.aws_subnet.application[*].id
  }
  provider = aws.eu_west_1
}
