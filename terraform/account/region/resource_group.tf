resource "aws_resourcegroups_group" "account" {
  name        = "${data.aws_default_tags.current.tags.account-name}-account-${data.aws_region.current.name}"
  description = "Account level resources for ${data.aws_region.current.name}"

  resource_query {
    query = local.account_resource_group_query
    type  = "TAG_FILTERS_1_0"
  }
  provider = aws.region
}

locals {
  account_resource_group_query = jsonencode({
    ResourceTypeFilters = [
      "AWS::AllSupported"
    ],
    TagFilters = [
      {
        Key    = "account-name",
        Values = [data.aws_default_tags.current.tags.account-name]
      }
    ]
  })
}
