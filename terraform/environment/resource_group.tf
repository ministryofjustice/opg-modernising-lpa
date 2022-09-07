resource "aws_resourcegroups_group" "environment_eu_west_1" {
  name        = "${local.default_tags.environment-name}-environment-eu-west-1"
  description = "Environment level eu-west-1 resources"

  resource_query {
    query = local.environment_resource_group_query
    type  = "TAG_FILTERS_1_0"
  }
  provider = aws.eu_west_1
}

resource "aws_resourcegroups_group" "environment_eu_west_2" {
  name        = "${local.default_tags.environment-name}-environment-eu-west-2"
  description = "Environment level eu-west-2 resources"

  resource_query {
    query = local.environment_resource_group_query
    type  = "TAG_FILTERS_1_0"
  }
  provider = aws.eu_west_2
}

resource "aws_resourcegroups_group" "environment_global" {
  name        = "${local.default_tags.environment-name}-environment-global"
  description = "Environment level global resources"

  resource_query {
    query = local.environment_resource_group_query
    type  = "TAG_FILTERS_1_0"
  }
  provider = aws.global
}

locals {
  environment_resource_group_query = jsonencode({
    ResourceTypeFilters = [
      "AWS::AllSupported"
    ],
    TagFilters = [
      {
        Key    = "environment-name",
        Values = [local.default_tags.environment-name]
      }
    ]
  })
}

output "resource_group_arns" {
  value = [
    aws_resourcegroups_group.environment_eu_west_1.arn,
    aws_resourcegroups_group.environment_eu_west_2.arn,
    aws_resourcegroups_group.environment_global.arn
  ]
}
