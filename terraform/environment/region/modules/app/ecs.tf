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
    update = "4m"
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
  name     = "${data.aws_default_tags.current.tags.environment-name}-${data.aws_region.current.name}-app-task-role"
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

data "aws_kms_alias" "reduced_fees_uploads_s3_encryption" {
  name     = "alias/${data.aws_default_tags.current.tags.application}_reduced_fees_uploads_s3_encryption"
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

locals {
  policy_region_prefix = lower(replace(data.aws_region.current.name, "-", ""))
}

data "aws_iam_policy_document" "task_role_access_policy" {
  policy_id = "${local.policy_region_prefix}task_role_access_policy"
  statement {
    sid    = "${local.policy_region_prefix}XrayAccess"
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
    sid    = "${local.policy_region_prefix}EcsDecryptAccess"
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
    sid    = "${local.policy_region_prefix}DynamoDBEncryptionAccess"
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
    sid    = "${local.policy_region_prefix}ReducedFeesUploadsEncryptionAccess"
    effect = "Allow"

    actions = [
      "kms:Encrypt",
      "kms:Decrypt",
      "kms:GenerateDataKey",
    ]

    resources = [
      data.aws_kms_alias.reduced_fees_uploads_s3_encryption.target_key_arn,
    ]
  }

  statement {
    sid    = "${local.policy_region_prefix}EcsSecretAccess"
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
    ]
  }

  statement {
    sid = "${local.policy_region_prefix}Allow"

    actions = [
      "dynamodb:BatchGetItem",
      "dynamodb:DeleteItem",
      "dynamodb:GetItem",
      "dynamodb:PutItem",
      "dynamodb:Query",
      "dynamodb:Scan",
      "dynamodb:UpdateItem",
    ]

    resources = [
      var.lpas_table.arn,
      "${var.lpas_table.arn}/index/*",
    ]
  }

  statement {
    sid    = "allowApiAccess"
    effect = "Allow"
    actions = [
      "execute-api:Invoke",
    ]
    resources = var.app_allowed_api_arns
  }

  statement {
    sid    = "uploadsS3BucketAccess"
    effect = "Allow"
    actions = [
      "s3:PutObject",
      "s3:PutObjectTagging",
      "s3:DeleteObject",
    ]
    resources = [
      "${var.uploads_s3_bucket.bucket_arn}/*",
    ]
  }

  statement {
    sid    = "${local.policy_region_prefix}EventbridgeAccess"
    effect = "Allow"

    actions = [
      "events:PutEvents"
    ]

    resources = [
      var.event_bus.arn
    ]
  }

  provider = aws.region
}


locals {
  app_url = "https://${data.aws_default_tags.current.tags.environment-name}.app.modernising.opg.service.justice.gov.uk"

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
          value = var.mock_onelogin_enabled ? "https://${data.aws_default_tags.current.tags.environment-name}-mock-onelogin.app.modernising.opg.service.justice.gov.uk" : "https://oidc.integration.account.gov.uk"
        },
        {
          name  = "MOCK_IDENTITY_PUBLIC_KEY",
          value = var.mock_onelogin_enabled ? "LS0tLS1CRUdJTiBQVUJMSUMgS0VZLS0tLS0KTUZrd0V3WUhLb1pJemowQ0FRWUlLb1pJemowREFRY0RRZ0FFSlEyVmtpZWtzNW9rSTIxY1Jma0FhOXVxN0t4TQo2bTJqWllCeHBybFVXQlpDRWZ4cTI3cFV0Qzd5aXplVlRiZUVqUnlJaStYalhPQjFBbDhPbHFtaXJnPT0KLS0tLS1FTkQgUFVCTElDIEtFWS0tLS0tCg==" : ""
        },
        {
          name  = "APP_PUBLIC_URL",
          value = var.app_env_vars.app_public_url == "" ? local.app_url : var.app_env_vars.app_public_url
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
          name  = "UPLOADS_S3_BUCKET_NAME",
          value = var.uploads_s3_bucket.bucket_name
        },
        {
          name  = "GOVUK_PAY_BASE_URL",
          value = "https://publicapi.payments.service.gov.uk"
        },
        {
          name  = "GOVUK_NOTIFY_BASE_URL",
          value = "https://api.notifications.service.gov.uk"
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
        {
          name  = "UID_BASE_URL",
          value = var.uid_base_url
        },
        {
          name  = "ONELOGIN_URL",
          value = var.app_env_vars.onelogin_url
        },
        {
          name  = "EVENT_BUS_NAME",
          value = var.event_bus.name
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
