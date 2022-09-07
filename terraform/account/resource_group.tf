resource "aws_resourcegroups_group" "account_eu_west_1" {
  name        = "${local.default_tags.account-name}-account-eu-west-1"
  description = "Account level eu-west-1 resources"

  resource_query {
    query = local.account_resource_group_query
    type  = "TAG_FILTERS_1_0"
  }
  provider = aws.eu_west_1
}

resource "aws_resourcegroups_group" "account_eu_west_2" {
  name        = "${local.default_tags.account-name}-account-eu-west-2"
  description = "Account level eu-west-2 resources"

  resource_query {
    query = local.account_resource_group_query
    type  = "TAG_FILTERS_1_0"
  }
  provider = aws.eu_west_2
}

resource "aws_resourcegroups_group" "account_global" {
  name        = "${local.default_tags.account-name}-account-global"
  description = "Account level global resources"

  resource_query {
    query = local.account_resource_group_query
    type  = "TAG_FILTERS_1_0"
  }
  provider = aws.global
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
