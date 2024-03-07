data "aws_vpc_endpoint" "opensearch" {
  tags = {
    Name = "opensearch-eu-west-1"
  }
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
    AWSOwnedKey = true
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
      Principal = [module.global.iam_roles.app_ecs_task_role.arn]
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
