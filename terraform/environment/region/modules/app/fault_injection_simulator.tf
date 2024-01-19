# Create encrypted logging for fault injection experiments

data "aws_kms_alias" "cloudwatch_application_logs_encryption" {
  name     = "alias/${data.aws_default_tags.current.tags.application}_cloudwatch_application_logs_encryption"
  provider = aws.region
}

resource "aws_cloudwatch_log_group" "fis_app_ecs_tasks" {
  name              = "fis/app-ecs-tasks-experiment-${data.aws_default_tags.current.tags.environment-name}"
  retention_in_days = 7
  kms_key_id        = data.aws_kms_alias.cloudwatch_application_logs_encryption.target_key_arn
  provider          = aws.region
}

# Add resource policy to allow FIS or the FIS role to write logs - not working
data "aws_iam_policy_document" "fis_app_ecs_tasks" {
  provider  = aws.region
  policy_id = "fis_app_ecs_tasks"
  statement {
    actions = [
      "logs:CreateLogDelivery",
      "logs:DescribeLogGroups",
      "logs:CreateLogStream",
      "logs:PutLogEvents",
      "logs:DescribeResourcePolicies",
    ]

    resources = [
      "arn:aws:logs:*:*:log-group:/fis/*",
      "arn:aws:logs:${data.aws_region.current.name}:${data.aws_caller_identity.current.account_id}:log-group:fis/app-ecs-tasks-experiment-936mlpab157:*"
    ]

    principals {
      identifiers = [data.aws_caller_identity.current.account_id]
      type        = "AWS"
    }
  }
}

resource "aws_cloudwatch_log_resource_policy" "fis_app_ecs_tasks" {
  provider        = aws.region
  policy_document = data.aws_iam_policy_document.fis_app_ecs_tasks.json
  policy_name     = "fis_app_ecs_tasks"
}

# Add log encryption permissions to the FIS role

data "aws_iam_policy_document" "fis_role_log_encryption" {
  provider  = aws.region
  policy_id = "fis_role_log_encryption"
  statement {
    actions = [
      "kms:Encrypt",
      "kms:GenerateDataKey",
    ]

    resources = [
      data.aws_kms_alias.cloudwatch_application_logs_encryption.target_key_arn
    ]
  }
}

resource "aws_iam_role_policy" "fis_role_log_encryption" {
  provider = aws.region
  name     = "fis_role_log_encryption"
  role     = var.fault_injection_simulator_role.name
  policy   = data.aws_iam_policy_document.fis_role_log_encryption.json
}

# Create experiment template for ECS tasks

resource "aws_fis_experiment_template" "ecs_app" {
  count       = data.aws_default_tags.current.tags.environment-name == "production" ? 0 : 1
  provider    = aws.region
  description = "Run ECS task experiments for the app service"
  role_arn    = var.fault_injection_simulator_role.arn
  tags = {
    Name = "${data.aws_default_tags.current.tags.environment-name} - APP ECS Task Experiments"
  }

  action {
    action_id   = "aws:ecs:task-cpu-stress"
    description = null
    name        = "cpu_stress_100_percent"
    parameter {
      key   = "duration"
      value = "PT5M"
    }
    target {
      key   = "Tasks"
      value = "app-ecs-tasks-${data.aws_default_tags.current.tags.environment-name}"
    }
  }

  stop_condition {
    source = "none"
    value  = null
  }

  # log_configuration {
  #   log_schema_version = 1

  #   cloudwatch_logs_configuration {
  #     log_group_arn = "${aws_cloudwatch_log_group.fis_app_ecs_tasks.arn}:*" # tfsec:ignore:aws-cloudwatch-log-group-wildcard
  #   }
  # }

  target {
    name = "app-ecs-tasks-${data.aws_default_tags.current.tags.environment-name}"
    resource_tag {
      key   = "environment-name"
      value = data.aws_default_tags.current.tags.environment-name
    }
    resource_type  = "aws:ecs:task"
    selection_mode = "ALL"
  }
}

# Create ECS task definition for ssm agent, used to run experiments

locals {
  amazon_ssm_agent = jsonencode(
    {
      name         = "amazon-ssm-agent",
      image        = "public.ecr.aws/amazon-ssm-agent/amazon-ssm-agent:latest",
      cpu          = 0,
      links        = [],
      portMappings = [],
      essential    = false,
      entryPoint   = [],
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
