resource "aws_ecs_service" "app" {
  name                  = "app"
  cluster               = var.ecs_cluster
  task_definition       = aws_ecs_task_definition.app.arn
  desired_count         = var.ecs_service_desired_count
  platform_version      = "1.4.0"
  wait_for_steady_state = true
  propagate_tags        = "SERVICE"

  capacity_provider_strategy {
    capacity_provider = var.ecs_capacity_provider
    weight            = 100
  }

  network_configuration {
    security_groups  = [aws_security_group.app_ecs_service.id]
    subnets          = var.network.application_subnets
    assign_public_ip = false
  }

  load_balancer {
    target_group_arn = aws_lb_target_group.app.arn
    container_name   = "app"
    container_port   = var.container_port
  }

  lifecycle {
    create_before_destroy = true
  }

  timeouts {
    create = "7m"
    update = "7m"
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

resource "aws_security_group_rule" "app_ecs_service_ingress" {
  description              = "Allow Port 80 ingress from the application load balancer"
  type                     = "ingress"
  from_port                = 80
  to_port                  = var.container_port
  protocol                 = "tcp"
  security_group_id        = aws_security_group.app_ecs_service.id
  source_security_group_id = aws_security_group.app_loadbalancer.id
  lifecycle {
    create_before_destroy = true
  }
  provider = aws.region
}

resource "aws_security_group_rule" "app_ecs_service_egress" {
  description       = "Allow any egress from service"
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
  container_definitions    = "[${local.app}, ${local.aws_otel_collector}]"
  task_role_arn            = var.ecs_task_role.arn
  execution_role_arn       = var.ecs_execution_role.arn
  provider                 = aws.region
}

resource "aws_iam_role_policy" "app_task_role" {
  name     = "${data.aws_default_tags.current.tags.environment-name}-app-task-role"
  policy   = data.aws_iam_policy_document.task_role_access_policy.json
  role     = var.ecs_task_role.name
  provider = aws.region
}

data "aws_kms_alias" "secrets_manager_secret_encryption_key" {
  name     = "alias/${data.aws_default_tags.current.tags.application}_secrets_manager_secret_encryption_key"
  provider = aws.region
}

data "aws_kms_alias" "dynamodb_encryption_key" {
  name     = "alias/${data.aws_default_tags.current.tags.application}_dynamodb_encryption"
  provider = aws.region
}

data "aws_secretsmanager_secret" "private_jwt_key" {
  name     = "private-jwt-key-base64"
  provider = aws.region
}

data "aws_secretsmanager_secret" "gov_uk_onelogin_identity_public_key" {
  name     = "gov-uk-onelogin-identity-public-key"
  provider = aws.region
}

data "aws_secretsmanager_secret" "cookie_session_keys" {
  name     = "cookie-session-keys"
  provider = aws.region
}

data "aws_secretsmanager_secret" "gov_uk_pay_api_key" {
  name     = "gov-uk-pay-api-key"
  provider = aws.region
}

data "aws_secretsmanager_secret" "yoti_private_key" {
  name     = "yoti-private-key"
  provider = aws.region
}

data "aws_secretsmanager_secret" "gov_uk_notify_api_key" {
  name     = "gov-uk-notify-api-key"
  provider = aws.region
}

data "aws_secretsmanager_secret" "os_postcode_lookup_api_key" {
  name     = "os-postcode-lookup-api-key"
  provider = aws.region
}

data "aws_secretsmanager_secret" "rum_monitor_identity_pool_id" {
  name     = "rum-monitor-identity-pool-id-${data.aws_region.current.name}"
  provider = aws.region
}


data "aws_iam_policy_document" "task_role_access_policy" {
  statement {
    sid    = "XrayAccess"
    effect = "Allow"

    actions = [
      "xray:PutTraceSegments",
      "xray:PutTelemetryRecords",
      "xray:GetSamplingRules",
      "xray:GetSamplingTargets",
      "xray:GetSamplingStatisticSummaries",
    ]

    resources = ["*"]
  }

  statement {
    sid    = "EcsDecryptAccess"
    effect = "Allow"

    actions = [
      "kms:Decrypt",
      "kms:GenerateDataKey",
    ]

    resources = [
      data.aws_kms_alias.secrets_manager_secret_encryption_key.target_key_arn,
    ]
  }

  statement {
    sid    = "DynamoDBEncryptionAccess"
    effect = "Allow"

    actions = [
      "kms:Encrypt",
      "kms:Decrypt",
      "kms:GenerateDataKey",
    ]

    resources = [
      data.aws_kms_alias.dynamodb_encryption_key.target_key_arn,
    ]
  }

  statement {
    sid    = "EcsSecretAccess"
    effect = "Allow"

    actions = [
      "secretsmanager:GetSecretValue",
      "secretsmanager:DescribeSecret",
    ]

    resources = [
      data.aws_secretsmanager_secret.cookie_session_keys.arn,
      data.aws_secretsmanager_secret.gov_uk_notify_api_key.arn,
      data.aws_secretsmanager_secret.gov_uk_onelogin_identity_public_key.arn,
      data.aws_secretsmanager_secret.gov_uk_pay_api_key.arn,
      data.aws_secretsmanager_secret.os_postcode_lookup_api_key.arn,
      data.aws_secretsmanager_secret.private_jwt_key.arn,
      data.aws_secretsmanager_secret.yoti_private_key.arn,
    ]
  }

  statement {
    sid = "Allow"

    actions = ["dynamodb:*"]

    resources = [
      var.lpas_table.arn,
      "${var.lpas_table.arn}/index/*",
    ]
  }

  provider = aws.region
}


locals {
  app = jsonencode(
    {
      cpu                    = 1,
      essential              = true,
      image                  = "${var.app_service_repository_url}:${var.app_service_container_version}",
      mountPoints            = [],
      readonlyRootFilesystem = true
      name                   = "app",
      portMappings = [
        {
          containerPort = var.container_port,
          hostPort      = var.container_port,
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
      secrets = [
        {
          name      = "AWS_RUM_IDENTITY_POOL_ID",
          valueFrom = data.aws_secretsmanager_secret.rum_monitor_identity_pool_id.arn
        },
        {
          name      = "AWS_RUM_APPLICATION_ID",
          valueFrom = var.rum_monitor_application_id_secretsmanager_secret_arn
        }
      ],
      environment = [
        {
          name  = "LOGGING_LEVEL",
          value = tostring(100)
        },
        {
          name  = "APP_PORT",
          value = tostring(var.container_port)
        },
        {
          name  = "CLIENT_ID",
          value = "37iOvkzc5BIRKsFSu5l3reZmFlA"
        },
        {
          name  = "ISSUER",
          value = "https://oidc.integration.account.gov.uk"
        },
        {
          name  = "APP_PUBLIC_URL",
          value = var.app_env_vars.app_public_url == "" ? "https://${local.dev_app_fqdn}" : var.app_env_vars.app_public_url
        },
        {
          # this is not the final value, but will allow signin to be tested while the real redirectURL is changed
          name  = "AUTH_REDIRECT_BASE_URL",
          value = var.app_env_vars.auth_redirect_base_url
        },
        {
          name  = "DYNAMODB_TABLE_LPAS",
          value = var.lpas_table.name
        },
        {
          name  = "GOVUK_PAY_BASE_URL",
          value = "https://publicapi.payments.service.gov.uk"
        },
        {
          name  = "YOTI_CLIENT_SDK_ID",
          value = var.app_env_vars.yoti_client_sdk_id
        },
        {
          name  = "YOTI_SCENARIO_ID",
          value = var.app_env_vars.yoti_scenario_id
        },
        {
          name = "YOTI_CERTIFICATE_PROVIDER_SCENARIO_ID",
          value = var.app_env_vars.yoti_certificate_provider_scenario_id,
        },
        {
          name  = "YOTI_SANDBOX",
          value = var.app_env_vars.yoti_sandbox
        },
        {
          name  = "ORDNANCE_SURVEY_BASE_URL",
          value = "https://api.os.uk"
        },
        {
          name  = "GOVUK_NOTIFY_IS_PRODUCTION",
          value = var.app_env_vars.notify_is_production
        },
        {
          name  = "XRAY_ENABLED",
          value = "1"
        },
        {
          name  = "AWS_RUM_GUEST_ROLE_ARN",
          value = var.aws_rum_guest_role_arn
        },
        {
          name  = "AWS_RUM_ENDPOINT",
          value = "https://dataplane.rum.${data.aws_region.current.name}.amazonaws.com"
        },
        {
          name  = "AWS_RUM_APPLICATION_REGION",
          value = data.aws_region.current.name
        },
      ]
    }
  )

  aws_otel_collector = jsonencode(
    {
      cpu                    = 0,
      essential              = true,
      image                  = "public.ecr.aws/aws-observability/aws-otel-collector:v0.21.0",
      mountPoints            = [],
      readonlyRootFilesystem = true
      name                   = "aws-otel-collector",
      command = [
        "--config=/etc/ecs/ecs-default-config.yaml"
      ],
      portMappings = [],
      volumesFrom  = [],
      logConfiguration = {
        logDriver = "awslogs",
        options = {
          awslogs-group         = var.ecs_application_log_group_name,
          awslogs-region        = data.aws_region.current.name,
          awslogs-stream-prefix = "${data.aws_default_tags.current.tags.environment-name}.otel.app"
        }
      },
      environment = []
  })
}
