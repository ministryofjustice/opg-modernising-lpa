data "aws_opensearchserverless_collection" "lpas_collection" {
  name     = "shared-collection-${local.environment.account_name}"
  provider = aws.eu_west_1
}

resource "aws_opensearchserverless_access_policy" "app" {
  name        = "app-${local.environment.account_name}"
  type        = "data"
  description = "allow index and collection access"
  policy = jsonencode([
    {
      Rules = [
        {
          ResourceType = "index",
          Resource     = ["index/collection-${local.environment.account_name}/*"],
          Permission   = ["aoss:*"]
        },
        {
          ResourceType = "collection",
          Resource     = ["collection/collection-${local.environment.account_name}"],
          Permission   = ["aoss:*"]
        }
      ],
      Principal = [
        module.global.iam_roles.app_ecs_task_role.arn,
      ]
    }
  ])
  provider = aws.eu_west_1
}

resource "aws_opensearchserverless_access_policy" "event_received" {
  name        = "event-received-${local.environment.account_name}"
  type        = "data"
  description = "allow index and collection access"
  policy = jsonencode([
    {
      Rules = [
        {
          ResourceType = "index",
          Resource     = ["index/collection-${local.environment.account_name}/*"],
          Permission   = ["aoss:*"]
        },
        {
          ResourceType = "collection",
          Resource     = ["collection/collection-${local.environment.account_name}"],
          Permission   = ["aoss:*"]
        }
      ],
      Principal = [module.global.iam_roles.event_received_lambda.arn]
    }
  ])
  provider = aws.eu_west_1
}
