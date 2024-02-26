data "aws_default_tags" "current" {
  provider = aws.eu_west_1
}

# data "aws_caller_identity" "current" {
#   provider = aws.eu_west_1
# }

data "aws_vpc" "main" {
  filter {
    name   = "tag:application"
    values = [data.aws_default_tags.current.tags.application]
  }
  provider = aws.eu_west_1
}

data "aws_availability_zones" "available" {
  state    = "available"
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

resource "aws_opensearchserverless_security_policy" "lpas_collection_encryption_policy" {
  name        = "policy-${local.environment_name}"
  type        = "encryption"
  description = "encryption policy for lpas-collection"
  policy = jsonencode({
    Rules = [
      {
        Resource = ["collection/lpas-collection"],
        ResourceType = "collection"
      }
    ],
    AWSOwnedKey = true
  })
  provider = aws.eu_west_1
}

resource "aws_opensearchserverless_collection" "lpas_collection" {
  name = "collection-${local.environment_name}"
  depends_on = [aws_opensearchserverless_security_policy.lpas_collection_encryption_policy]
  provider = aws.eu_west_1
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
          Resource = ["collection/lpas-collection"]
        }
      ],
      AllowFromPublic = false,
      SourceVPCEs = [
        aws_opensearchserverless_vpc_endpoint.lpas_collection_vpc_endpoint.id
      ]
    },
    # {
    #   Description = "Public access for dashboards",
    #   Rules = [
    #     {
    #       ResourceType = "dashboard"
    #       Resource = ["collection/lpas-collection"]
    #     }
    #   ],
    #   AllowFromPublic = true
    # }
  ])
  provider = aws.eu_west_1
}

resource "aws_opensearchserverless_vpc_endpoint" "lpas_collection_vpc_endpoint" {
  name               = "endpoint-${local.environment_name}"
  vpc_id             = data.aws_vpc.main.id
  subnet_ids         = data.aws_subnet.application[*].id
  # security_group_ids = [data.aws_security_group.security_group.id]
  provider = aws.eu_west_1
}

resource "aws_opensearchserverless_access_policy" "lpas_collection_data_access_policy" {
  name        = "policy-${local.environment_name}"
  type        = "data"
  description = "allow index and collection access"
  policy = jsonencode([
    {
      Rules = [
        {
          ResourceType = "index",
          Resource = ["index/lpas-collection/*"],
          Permission = ["aoss:*"]
        },
        {
          ResourceType = "collection",
          Resource = ["collection/lpas-collection"],
          Permission = ["aoss:*"]
        }
      ],
      Principal = ["*"]
      # Principal = [data.aws_caller_identity.current.arn]
    }
  ])
  provider = aws.eu_west_1
}
