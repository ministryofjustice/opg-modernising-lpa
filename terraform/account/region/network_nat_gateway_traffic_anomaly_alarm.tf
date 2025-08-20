data "aws_nat_gateways" "ngws" {
  vpc_id = data.aws_vpc.main.id

  filter {
    name   = "state"
    values = ["available"]
  }
  provider = aws.region
  depends_on = [
    module.network
  ]
}

resource "aws_cloudwatch_metric_alarm" "nat_traffic_increase_anomaly_detection" {
  count                     = length(data.aws_nat_gateways.ngws.ids)
  alarm_name                = "nat-gateway-inbound-traffic-increase-anomaly-${count.index}"
  comparison_operator       = "GreaterThanUpperThreshold"
  evaluation_periods        = 2
  threshold_metric_id       = "ad${count.index}"
  alarm_description         = "This metric monitors NAT Gateway traffic into the VPC"
  insufficient_data_actions = []

  metric_query {
    id          = "ad${count.index}"
    return_data = true
    expression  = "ANOMALY_DETECTION_BAND(m${count.index}, 4)"
    label       = "AWS NAT Gateway ${tolist(data.aws_nat_gateways.ngws.ids)[count.index]} BytesOutToSource (Expected)"
  }

  metric_query {
    id          = "m${count.index}"
    return_data = true
    metric {
      metric_name = "BytesOutToSource"
      namespace   = "AWS/NATGateway"
      period      = 120
      stat        = "Average"
      unit        = "Bytes"

      dimensions = {
        NatGatewayId = tolist(data.aws_nat_gateways.ngws.ids)[count.index]
      }
    }
  }
  provider = aws.region
}
