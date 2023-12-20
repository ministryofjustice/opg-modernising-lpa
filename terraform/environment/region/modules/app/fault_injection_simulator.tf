data "aws_kms_alias" "cloudwatch_application_logs_encryption" {
  name     = "alias/${data.aws_default_tags.current.tags.application}_cloudwatch_application_logs_encryption"
  provider = aws.region
}

resource "aws_cloudwatch_log_group" "fis_app_ecs_tasks" {
  name              = "/aws/fis/app-ecs-tasks-experiment-${data.aws_default_tags.current.tags.environment-name}"
  retention_in_days = 7
  # kms_key_id        = data.aws_kms_alias.cloudwatch_application_logs_encryption.target_key_arn
  provider = aws.region
}

resource "aws_fis_experiment_template" "ecs_app" {
  count       = data.aws_default_tags.current.tags.environment-name == "production" ? 0 : 1
  provider    = aws.region
  description = "Run ECS task experiments for the app service"
  role_arn    = var.fault_injection_simulator_role_arn
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

  target {
    name = "app-ecs-tasks-${data.aws_default_tags.current.tags.environment-name}"
    resource_arns = [
      "arn:aws:ecs:eu-west-1:653761790766:task/936mlpab157-eu-west-1/1b1aedd43f94458d9ef3475479e19169",
      # aws_ecs_task_definition.mock_onelogin.arn,
    ]
    # parameters = {
    #   "cluster" : var.ecs_cluster,
    #   "service" : aws_ecs_service.app.name,
    # }
    # resource_tag {
    #   key   = "aws:ecs:service"
    #   value = aws_ecs_service.app.name
    # }
    resource_type  = "aws:ecs:task"
    selection_mode = "ALL"
  }
}
