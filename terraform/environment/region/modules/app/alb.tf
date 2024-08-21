resource "aws_lb_target_group" "app" {
  name                 = "${data.aws_default_tags.current.tags.environment-name}-app"
  port                 = 80
  protocol             = "HTTP"
  target_type          = "ip"
  vpc_id               = var.network.vpc_id
  deregistration_delay = 0
  depends_on           = [aws_lb.app]

  health_check {
    enabled = true
    path    = "/health-check/service"
  }

  provider = aws.region
}

resource "aws_lb" "app" {
  name                       = "${data.aws_default_tags.current.tags.environment-name}-app"
  internal                   = false #tfsec:ignore:AWS005 - public alb
  load_balancer_type         = "application"
  drop_invalid_header_fields = true
  subnets                    = var.network.public_subnets
  enable_deletion_protection = var.alb_deletion_protection_enabled
  security_groups            = [aws_security_group.app_loadbalancer.id]

  access_logs {
    bucket  = data.aws_s3_bucket.access_log.bucket
    prefix  = "app-${data.aws_default_tags.current.tags.environment-name}"
    enabled = true
  }
  provider = aws.region
}

resource "aws_lb_listener" "app_loadbalancer_http_redirect" {
  load_balancer_arn = aws_lb.app.arn
  port              = "80"
  protocol          = "HTTP"

  default_action {
    type = "redirect"

    redirect {
      port        = 443
      protocol    = "HTTPS"
      status_code = "HTTP_301"
    }
  }
  provider = aws.region
}

locals {
  dev_wildcard = data.aws_default_tags.current.tags.environment-name == "production" ? "" : "*."
}

data "aws_acm_certificate" "certificate_app" {
  domain   = "${local.dev_wildcard}app.modernising.opg.service.justice.gov.uk"
  provider = aws.region
}

resource "aws_lb_listener" "app_loadbalancer" {
  load_balancer_arn = aws_lb.app.arn
  port              = "443"
  protocol          = "HTTPS"
  ssl_policy        = "ELBSecurityPolicy-FS-1-2-2019-08"
  certificate_arn   = data.aws_acm_certificate.certificate_app.arn

  default_action {
    target_group_arn = aws_lb_target_group.app.arn
    type             = "forward"
  }
  provider = aws.region
}

resource "aws_lb_listener_rule" "app_maintenance" {
  listener_arn = aws_lb_listener.app_loadbalancer.arn
  priority     = 101 # Specifically set so that maintenance mode scripts can locate the correct rule to modify
  action {
    type = "redirect"

    redirect {
      host        = "maintenance.opg.service.justice.gov.uk"
      path        = "/en-gb/modernised-make-a-lasting-power-of-attorney" # temporarily the english make a lasting power of attorney maintenance page
      query       = ""
      port        = "443"
      protocol    = "HTTPS"
      status_code = "HTTP_302"
    }
  }
  condition {
    path_pattern {
      values = ["/maintenance"]
    }
  }
  lifecycle {
    ignore_changes = [
      # Ignore changes to the condition as this is modified by a script
      # when putting the service into maintenance mode.
      condition,
    ]
  }
  provider = aws.region
}

resource "aws_lb_listener_rule" "app_maintenance_welsh" {
  listener_arn = aws_lb_listener.app_loadbalancer.arn
  priority     = 100 # Specifically set so that maintenance mode scripts can locate the correct rule to modify
  action {
    type = "redirect"

    redirect {
      host        = "maintenance.opg.service.justice.gov.uk"
      path        = "/cy/moderneiddio-gwneud-atwrneiaeth-arhosol" # temporarily the welsh use a lasting power of attorney maintenance page
      query       = ""
      port        = "443"
      protocol    = "HTTPS"
      status_code = "HTTP_302"
    }
  }
  condition {
    path_pattern {
      values = ["/cy/maintenance"]
    }
  }
  lifecycle {
    ignore_changes = [
      # Ignore changes to the condition as this is modified by a script
      # when putting the service into maintenance mode.
      condition,
    ]
  }
  provider = aws.region
}

resource "aws_lb_listener_certificate" "app_loadbalancer_live_service_certificate" {
  listener_arn    = aws_lb_listener.app_loadbalancer.arn
  certificate_arn = data.aws_acm_certificate.certificate_app.arn
  provider        = aws.region
}

resource "aws_security_group" "app_loadbalancer" {
  name_prefix = "${data.aws_default_tags.current.tags.environment-name}-app-loadbalancer"
  description = "app service application load balancer"
  vpc_id      = var.network.vpc_id
  lifecycle {
    create_before_destroy = true
  }
  provider = aws.region
}

data "aws_ip_ranges" "route53_healthchecks" {
  services = ["route53_healthchecks"]
  regions  = ["GLOBAL", "us-east-1", "eu-west-1", "ap-southeast-1"]
  provider = aws.region
}

resource "terraform_data" "ingress_allow_list_cidr" {
  input = var.ingress_allow_list_cidr
}

resource "aws_security_group_rule" "app_loadbalancer_port_80_redirect_ingress" {
  count             = var.public_access_enabled ? 0 : 1
  description       = "Port 80 ingress for redirection to port 443"
  type              = "ingress"
  from_port         = 80
  to_port           = 80
  protocol          = "tcp"
  cidr_blocks       = var.ingress_allow_list_cidr #tfsec:ignore:aws-vpc-no-public-ingress-sgr
  security_group_id = aws_security_group.app_loadbalancer.id
  lifecycle {
    replace_triggered_by = [
      terraform_data.ingress_allow_list_cidr
    ]
  }
  provider = aws.region
}

resource "aws_security_group_rule" "app_loadbalancer_ingress" {
  count             = var.public_access_enabled ? 0 : 1
  description       = "Port 443 ingress from the allow list to the application load balancer"
  type              = "ingress"
  from_port         = 443
  to_port           = 443
  protocol          = "tcp"
  cidr_blocks       = var.ingress_allow_list_cidr #tfsec:ignore:aws-vpc-no-public-ingress-sgr
  security_group_id = aws_security_group.app_loadbalancer.id
  lifecycle {
    replace_triggered_by = [
      terraform_data.ingress_allow_list_cidr
    ]
  }
  provider = aws.region
}

resource "aws_security_group_rule" "loadbalancer_ingress_route53_healthchecks" {
  description       = "Loadbalancer ingresss from Route53 healthchecks"
  type              = "ingress"
  protocol          = "tcp"
  from_port         = "443"
  to_port           = "443"
  cidr_blocks       = data.aws_ip_ranges.route53_healthchecks.cidr_blocks
  ipv6_cidr_blocks  = data.aws_ip_ranges.route53_healthchecks.ipv6_cidr_blocks
  security_group_id = aws_security_group.app_loadbalancer.id
  provider          = aws.region
}

resource "aws_security_group_rule" "app_loadbalancer_public_access_ingress" {
  count             = var.public_access_enabled ? 1 : 0
  description       = "Port 443 production public ingress to the application load balancer"
  type              = "ingress"
  from_port         = 443
  to_port           = 443
  protocol          = "tcp"
  cidr_blocks       = ["0.0.0.0/0"] #tfsec:ignore:aws-vpc-no-public-ingress-sgr - open ingress for production
  security_group_id = aws_security_group.app_loadbalancer.id
  provider          = aws.region
}

# this has a listener rule in the alb to redirect to :443
resource "aws_security_group_rule" "app_loadbalancer_public_access_ingress_port_80" {
  count             = var.public_access_enabled ? 1 : 0
  description       = "Port 80 production public ingress to the application load balancer"
  type              = "ingress"
  from_port         = 80
  to_port           = 80
  protocol          = "tcp"
  cidr_blocks       = ["0.0.0.0/0"] #tfsec:ignore:aws-vpc-no-public-ingress-sgr - open ingress for production
  security_group_id = aws_security_group.app_loadbalancer.id
  provider          = aws.region
}

resource "aws_security_group_rule" "app_loadbalancer_egress" {
  description       = "Allow any egress from service load balancer"
  type              = "egress"
  from_port         = 0
  to_port           = 0
  protocol          = "-1"
  cidr_blocks       = ["0.0.0.0/0"] #tfsec:ignore:aws-ec2-no-public-egress-sgr - open egress for load balancers
  security_group_id = aws_security_group.app_loadbalancer.id
  provider          = aws.region
}
