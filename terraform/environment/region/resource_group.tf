resource "aws_resourcegroups_group" "environment" {
  name        = "${data.aws_default_tags.current.tags.environment-name}-environment-${data.aws_region.current.name}"
  description = "Environment level eu-west-1 resources"

  resource_query {
    query = local.environment_resource_group_query
    type  = "TAG_FILTERS_1_0"
  }
  provider = aws.region
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
