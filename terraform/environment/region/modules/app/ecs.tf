resource "aws_ecs_service" "app" {
  name                  = "app"
  cluster               = var.ecs_cluster
  task_definition       = aws_ecs_task_definition.app.arn
  desired_count         = var.ecs_service_desired_count
  platform_version      = "1.4.0"
  wait_for_steady_state = true

  capacity_provider_strategy {
    capacity_provider = var.ecs_capacity_provider
    weight            = 100
  }

  network_configuration {
    security_groups  = [aws_security_group.app_ecs_service.id]
    subnets          = var.network.application_subnets
    assign_public_ip = false
  }

  # load_balancer {
  #   target_group_arn = aws_lb_target_group.app.arn
  #   container_name   = "app"
  #   container_port   = 80
  # }

  lifecycle {
    create_before_destroy = true
  }
  provider = aws.region
}

resource "aws_security_group" "app_ecs_service" {
  name_prefix = "${local.name_prefix}-ecs-service"
  description = "app service security group"
  vpc_id      = var.network.vpc_id
  lifecycle {
    create_before_destroy = true
  }
  provider = aws.region
}

# resource "aws_security_group_rule" "app_ecs_service_ingress" {
#   description              = "Allow Port 80 ingress from the applciation load balancer"
#   type                     = "ingress"
#   from_port                = 80
#   to_port                  = 80
#   protocol                 = "tcp"
#   security_group_id        = aws_security_group.app_ecs_service.id
#   source_security_group_id = aws_security_group.app_loadbalancer.id
#   lifecycle {
#     create_before_destroy = true
#   }
#   provider = aws.region
# }

resource "aws_security_group_rule" "app_ecs_service_egress" {
  description       = "Allow any egress from Use service"
  type              = "egress"
  from_port         = 0
  to_port           = 0
  protocol          = "-1"
  cidr_blocks       = ["0.0.0.0/0"] #tfsec:ignore:aws-ec2-no-public-egress-sgr - open egress for ECR access
  security_group_id = aws_security_group.app_ecs_service.id
  lifecycle {
    create_before_destroy = true
  }
  provider = aws.region
}

resource "aws_ecs_task_definition" "app" {
  family                   = local.name_prefix
  requires_compatibilities = ["FARGATE"]
  network_mode             = "awsvpc"
  cpu                      = 512
  memory                   = 1024
  container_definitions    = "[${local.app_app}]"
  task_role_arn            = var.ecs_task_role_arn
  execution_role_arn       = var.ecs_execution_role.arn
  provider                 = aws.region
}


locals {
  app_app = jsonencode(
    {
      cpu         = 1,
      essential   = true,
      image       = "${var.app_service_repository_url}:${var.app_service_container_version}",
      mountPoints = [],
      name        = "app",
      portMappings = [
        {
          containerPort = 80,
          hostPort      = 80,
          protocol      = "tcp"
        }
      ],
      volumesFrom = [],
      logConfiguration = {
        logDriver = "awslogs",
        options = {
          awslogs-group         = var.ecs_application_log_group_name,
          awslogs-region        = data.aws_region.current.name,
          awslogs-stream-prefix = data.aws_default_tags.current.tags.environment-name
        }
      },
      environment = [
        {
          name  = "LOGGING_LEVEL",
          value = tostring(100)
        },
        {
          name  = "APP_PORT",
          value = tostring(80)
        }
      ]
    }
  )
}
