module "network" {
  source                              = "github.com/ministryofjustice/opg-terraform-aws-firewalled-network?ref=v0.2.16"
  cidr                                = var.network_cidr_block
  enable_dns_hostnames                = true
  enable_dns_support                  = true
  default_security_group_ingress      = []
  default_security_group_egress       = []
  aws_networkfirewall_firewall_policy = aws_networkfirewall_firewall_policy.main
  providers = {
    aws = aws.region
  }
}

resource "aws_networkfirewall_firewall_policy" "main" {
  name = "main"

  firewall_policy {
    stateless_default_actions          = ["aws:forward_to_sfe"]
    stateless_fragment_default_actions = ["aws:forward_to_sfe"]

    stateful_engine_options {
      rule_order              = "DEFAULT_ACTION_ORDER"
      stream_exception_policy = "DROP"
    }
    stateful_rule_group_reference {
      resource_arn = aws_networkfirewall_rule_group.rule_file.arn
    }
  }
  provider = aws.region
}

resource "aws_networkfirewall_rule_group" "rule_file" {
  capacity = 100
  name     = "main-${replace(filebase64sha256("${path.module}/network_firewall_rules.rules"), "/[^[:alnum:]]/", "")}"
  type     = "STATEFUL"
  rules    = file("${path.module}/network_firewall_rules.rules")
  lifecycle {
    create_before_destroy = true
  }
  provider = aws.region
}

resource "aws_security_group" "lambda_egress" {
  name        = "lambda-egress-${data.aws_region.current.name}"
  vpc_id      = module.network.vpc.id
  description = "Shared security group lambda for outbound traffic"

  tags     = { "Name" = "lambda-egress-${data.aws_region.current.name}" }
  provider = aws.region
}

#tfsec:ignore:aws-ec2-no-public-egress-sgr egress is managed by AWS Network Firewall
resource "aws_security_group_rule" "lambda_egress" {
  type              = "egress"
  protocol          = "-1"
  from_port         = 0
  to_port           = 0
  cidr_blocks       = ["0.0.0.0/0"]
  security_group_id = aws_security_group.lambda_egress.id
  description       = "Outbound Lambda"
  provider          = aws.region
}
