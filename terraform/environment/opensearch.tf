resource "aws_opensearchserverless_security_policy" "lpas_collection_encryption_policy" {
  name        = "lpas-collection-encryption-policy"
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
}

resource "aws_opensearchserverless_collection" "lpas_collection" {
  name = "lpas-collection"
  depends_on = [aws_opensearchserverless_security_policy.lpas_collection_encryption_policy]
}

resource "aws_opensearchserverless_security_policy" "lpas_collection_network_policy" {
  name        = "lpas-collection-network-policy"
  type        = "network"
  description = "public access for dashboard, VPC access for collection endpoint"
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
        aws_opensearchserverless_vpc_endpoint.vpc_endpoint.id
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
}

resource "aws_opensearchserverless_vpc_endpoint" "lpas_collection_vpc_endpoint" {
  name               = "lpas-collection-vpc-endpoint"
  vpc_id             = aws_vpc.vpc.id
  subnet_ids         = [aws_subnet.subnet.id]
  security_group_ids = [aws_security_group.security_group.id]
}

resource "aws_opensearchserverless_access_policy" "lpas_collection_data_access_policy" {
  name        = "lpas-collection-data-access-policy"
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
      Principal = [data.aws_caller_identity.global.arn]
    }
  ])
}
