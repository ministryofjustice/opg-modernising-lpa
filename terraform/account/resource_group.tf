resource "aws_resourcegroups_group" "account" {
  name        = "account-${local.default_tags.account-name}"
  description = "Account level resources"

  resource_query {
    query = local.account_resource_group_query
  }
}

locals {
  account_resource_group_query = jsonencode({
    ResourceTypeFilters = [
      "AWS::AllSupported"
    ],
    TagFilters = [
      {
        Key    = "account-name",
        Values = [local.default_tags.account-name]
      }
    ]
  })
}
