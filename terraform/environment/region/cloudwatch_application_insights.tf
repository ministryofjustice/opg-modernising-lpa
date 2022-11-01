resource "aws_applicationinsights_application" "environment_eu_west_1" {
  resource_group_name = "${data.aws_default_tags.current.tags.environment-name}-environment-${data.aws_region.current.name}"
  provider            = aws.region
}
