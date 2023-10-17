data "aws_vpc" "main" {
  filter {
    name   = "tag:application"
    values = [data.aws_default_tags.current.tags.application]
  }
  provider = aws.region
}

data "aws_availability_zones" "available" {
  provider = aws.region
}

data "aws_subnet" "application" {
  count             = 3
  vpc_id            = data.aws_vpc.main.id
  availability_zone = data.aws_availability_zones.available.names[count.index]

  filter {
    name   = "tag:Name"
    values = ["application*"]
  }
  provider = aws.region
}

data "aws_subnet" "public" {
  count             = 3
  vpc_id            = data.aws_vpc.main.id
  availability_zone = data.aws_availability_zones.available.names[count.index]

  filter {
    name   = "tag:Name"
    values = ["public*"]
  }
  provider = aws.region
}

# data "aws_vpc_endpoint" "s3" {
#   vpc_id       = data.aws_vpc.main.id
#   service_name = "com.amazonaws.${data.aws_region.current.name}.s3"
#   provider     = aws.region
# }
