resource "aws_resourcegroups_group" "environment_global" {
  name        = "${data.aws_default_tags.current.tags.environment-name}-environment-global"
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
        Values = [data.aws_default_tags.current.tags.environment-name]
      }
    ]
  })
}

output "resource_group_arns" {
  value = [
    aws_resourcegroups_group.environment_global.arn
  ]
}
