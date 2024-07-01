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

  deployment_controller {
    type = "ECS"
  }

  deployment_circuit_breaker {
    enable   = true
    rollback = true
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
    ignore_changes = [
      desired_count
    ]
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
  family                   = "${local.name_prefix}-app"
  requires_compatibilities = ["FARGATE"]
  network_mode             = "awsvpc"
  cpu                      = 512
  memory                   = 1024
  runtime_platform {
    operating_system_family = "LINUX"
    cpu_architecture        = "X86_64"
  }
  container_definitions = var.fault_injection_experiments_enabled ? "[${local.app}, ${local.aws_otel_collector}, ${local.amazon_ssm_agent}]" : "[${local.app}, ${local.aws_otel_collector}]"
  task_role_arn         = var.ecs_task_role.arn
  execution_role_arn    = var.ecs_execution_role.arn
  provider              = aws.region
}

resource "aws_iam_role_policy" "app_task_role" {
  name     = "${data.aws_default_tags.current.tags.environment-name}-${data.aws_region.current.name}-app-task-role"
  policy   = var.fault_injection_experiments_enabled ? data.aws_iam_policy_document.combined.json : data.aws_iam_policy_document.task_role_access_policy.json
  role     = var.ecs_task_role.name
  provider = aws.region
}

data "aws_iam_policy_document" "combined" {
  source_policy_documents = [
    data.aws_iam_policy_document.task_role_access_policy.json,
    data.aws_iam_policy_document.ecs_task_role_fis_related_task_permissions.json
  ]
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

data "aws_kms_alias" "opensearch_encryption_key" {
  name     = "alias/${data.aws_default_tags.current.tags.application}-opensearch-encryption-key"
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

data "aws_secretsmanager_secret" "lpa_store_jwt_secret_key" {
  name     = "lpa-store-jwt-secret-key"
  provider = aws.region
}

data "aws_secretsmanager_secret" "rum_monitor_identity_pool_id" {
  name     = "rum-monitor-identity-pool-id-${data.aws_region.current.name}"
  provider = aws.region
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
    sid    = "${local.policy_region_prefix}OpensearchEncryptionAccess"
    effect = "Allow"

    actions = [
      "kms:Encrypt",
      "kms:Decrypt",
      "kms:GenerateDataKey",
    ]

    resources = [
      data.aws_kms_alias.opensearch_encryption_key.target_key_arn,
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
      data.aws_secretsmanager_secret.lpa_store_jwt_secret_key.arn,
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

  statement {
    sid    = "${local.policy_region_prefix}OpenSearchAccess"
    effect = "Allow"

    actions = [
      "aoss:APIAccessAll"
    ]

    resources = [
      var.search_collection_arn
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
          value = var.mock_onelogin_enabled ? "http://mock-onelogin.${data.aws_default_tags.current.tags.environment-name}.internal.modernising.ecs:8080" : "https://oidc.integration.account.gov.uk"
        },
        {
          name  = "MOCK_IDENTITY_PUBLIC_KEY",
          value = var.mock_onelogin_enabled ? "LS0tLS1CRUdJTiBQVUJMSUMgS0VZLS0tLS0KTUZrd0V3WUhLb1pJemowQ0FRWUlLb1pJemowREFRY0RRZ0FFSlEyVmtpZWtzNW9rSTIxY1Jma0FhOXVxN0t4TQo2bTJqWllCeHBybFVXQlpDRWZ4cTI3cFV0Qzd5aXplVlRiZUVqUnlJaStYalhPQjFBbDhPbHFtaXJnPT0KLS0tLS1FTkQgUFVCTElDIEtFWS0tLS0tCg==" : "" #pragma: allowlist secret
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
          value = var.mock_pay_enabled ? "http://mock-pay.${data.aws_default_tags.current.tags.environment-name}.internal.modernising.ecs:8080" : "https://publicapi.payments.service.gov.uk"
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
        {
          name  = "LPA_STORE_BASE_URL",
          value = var.lpa_store_base_url
        },
        {
          name  = "SEARCH_ENDPOINT",
          value = var.search_endpoint == null ? "" : var.search_endpoint
        },
        {
          name  = "SEARCH_INDEX_NAME",
          value = var.search_index_name
        },
        {
          name  = "DEV_MODE",
          value = var.app_env_vars.dev_mode
        },
        {
          name  = "SEARCH_INDEXING_DISABLED",
          value = "1"
        }
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

  amazon_ssm_agent = jsonencode(
    {
      name                   = "amazon-ssm-agent",
      image                  = "public.ecr.aws/amazon-ssm-agent/amazon-ssm-agent:latest",
      cpu                    = 0,
      links                  = [],
      portMappings           = [],
      essential              = false,
      entryPoint             = [],
      readonlyRootFilesystem = false
      command = [
        "/bin/bash",
        "-c",
        "set -e; yum upgrade -y; yum install jq procps awscli -y; term_handler() { echo \"Deleting SSM activation $ACTIVATION_ID\"; if ! aws ssm delete-activation --activation-id $ACTIVATION_ID --region $ECS_TASK_REGION; then echo \"SSM activation $ACTIVATION_ID failed to be deleted\" 1>&2; fi; MANAGED_INSTANCE_ID=$(jq -e -r .ManagedInstanceID /var/lib/amazon/ssm/registration); echo \"Deregistering SSM Managed Instance $MANAGED_INSTANCE_ID\"; if ! aws ssm deregister-managed-instance --instance-id $MANAGED_INSTANCE_ID --region $ECS_TASK_REGION; then echo \"SSM Managed Instance $MANAGED_INSTANCE_ID failed to be deregistered\" 1>&2; fi; kill -SIGTERM $SSM_AGENT_PID; }; trap term_handler SIGTERM SIGINT; if [[ -z $MANAGED_INSTANCE_ROLE_NAME ]]; then echo \"Environment variable MANAGED_INSTANCE_ROLE_NAME not set, exiting\" 1>&2; exit 1; fi; if ! ps ax | grep amazon-ssm-agent | grep -v grep > /dev/null; then if [[ -n $ECS_CONTAINER_METADATA_URI_V4 ]] ; then echo \"Found ECS Container Metadata, running activation with metadata\"; TASK_METADATA=$(curl \"$${ECS_CONTAINER_METADATA_URI_V4}/task\"); ECS_TASK_AVAILABILITY_ZONE=$(echo $TASK_METADATA | jq -e -r '.AvailabilityZone'); ECS_TASK_ARN=$(echo $TASK_METADATA | jq -e -r '.TaskARN'); ECS_TASK_REGION=$(echo $ECS_TASK_AVAILABILITY_ZONE | sed 's/.$//'); ECS_TASK_AVAILABILITY_ZONE_REGEX='^(af|ap|ca|cn|eu|me|sa|us|us-gov)-(central|north|(north(east|west))|south|south(east|west)|east|west)-[0-9]{1}[a-z]{1}$'; if ! [[ $ECS_TASK_AVAILABILITY_ZONE =~ $ECS_TASK_AVAILABILITY_ZONE_REGEX ]]; then echo \"Error extracting Availability Zone from ECS Container Metadata, exiting\" 1>&2; exit 1; fi; ECS_TASK_ARN_REGEX='^arn:(aws|aws-cn|aws-us-gov):ecs:[a-z0-9-]+:[0-9]{12}:task/[a-zA-Z0-9_-]+/[a-zA-Z0-9]+$'; if ! [[ $ECS_TASK_ARN =~ $ECS_TASK_ARN_REGEX ]]; then echo \"Error extracting Task ARN from ECS Container Metadata, exiting\" 1>&2; exit 1; fi; CREATE_ACTIVATION_OUTPUT=$(aws ssm create-activation --iam-role $MANAGED_INSTANCE_ROLE_NAME --tags Key=ECS_TASK_AVAILABILITY_ZONE,Value=$ECS_TASK_AVAILABILITY_ZONE Key=ECS_TASK_ARN,Value=$ECS_TASK_ARN Key=FAULT_INJECTION_SIDECAR,Value=true --region $ECS_TASK_REGION); ACTIVATION_CODE=$(echo $CREATE_ACTIVATION_OUTPUT | jq -e -r .ActivationCode); ACTIVATION_ID=$(echo $CREATE_ACTIVATION_OUTPUT | jq -e -r .ActivationId); if ! amazon-ssm-agent -register -code $ACTIVATION_CODE -id $ACTIVATION_ID -region $ECS_TASK_REGION; then echo \"Failed to register with AWS Systems Manager (SSM), exiting\" 1>&2; exit 1; fi; amazon-ssm-agent & SSM_AGENT_PID=$!; wait $SSM_AGENT_PID; else echo \"ECS Container Metadata not found, exiting\" 1>&2; exit 1; fi; else echo \"SSM agent is already running, exiting\" 1>&2; exit 1; fi"
      ],
      environment = [
        {
          name  = "MANAGED_INSTANCE_ROLE_NAME",
          value = "ssm-register-instance-${data.aws_default_tags.current.tags.environment-name}"
        }
      ],
      logConfiguration = {
        logDriver = "awslogs",
        options = {
          awslogs-group         = var.ecs_application_log_group_name,
          awslogs-region        = data.aws_region.current.name,
          awslogs-stream-prefix = "${data.aws_default_tags.current.tags.environment-name}.otel.app"
        }
      },
      environmentFiles      = [],
      mountPoints           = [],
      volumesFrom           = [],
      secrets               = [],
      dnsServers            = [],
      dnsSearchDomains      = [],
      extraHosts            = [],
      dockerSecurityOptions = [],
      dockerLabels          = {},
      ulimits               = [],
      systemControls        = []
  })
}

# Additional permissions for the ECS task role to run experiments

data "aws_iam_policy_document" "ecs_task_role_fis_related_task_permissions" {
  policy_id = "${local.policy_region_prefix}_fis_ecs_task_actions"
  statement {
    sid       = "AllowSSMCommands"
    effect    = "Allow"
    resources = ["*"] #tfsec:ignore:aws-iam-no-policy-wildcards
    actions = [
      "ssm:CreateActivation",
      "ssm:AddTagsToResource",
    ]
  }

  statement {
    sid       = "ManagedInstancePermissions"
    effect    = "Allow"
    resources = ["*"] #tfsec:ignore:aws-iam-no-policy-wildcards
    actions = [
      "ssm:DeleteActivation",
      "ssm:DeregisterManagedInstance",
    ]
  }

  statement {
    sid       = "AllowPassRole"
    effect    = "Allow"
    resources = ["*"] #tfsec:ignore:aws-iam-no-policy-wildcards
    actions = [
      "iam:PassRole",
    ]
  }
}
