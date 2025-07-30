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

data "aws_nat_gateways" "main" {
  vpc_id = data.aws_vpc.main.id

  filter {
    name   = "state"
    values = ["available"]
  }
  provider = aws.region
}

data "aws_nat_gateway" "main" {
  count    = length(data.aws_nat_gateways.main.ids)
  id       = tolist(data.aws_nat_gateways.main.ids)[count.index]
  provider = aws.region
}
