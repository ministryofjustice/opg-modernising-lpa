data "aws_kms_alias" "cloudwatch_application_logs_encryption" {
  name     = "alias/${data.aws_default_tags.current.tags.application}_cloudwatch_application_logs_encryption"
  provider = aws.region
}

resource "aws_cloudwatch_log_group" "fis_app_ecs_tasks" {
  name              = "/aws/fis/app-ecs-tasks-experiment-${data.aws_default_tags.current.tags.environment-name}"
  retention_in_days = 7
  kms_key_id        = data.aws_kms_alias.cloudwatch_application_logs_encryption.target_key_arn
  provider          = aws.region
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
    action_id   = "aws:ecs:stop-task"
    description = null
    name        = "stop_task"
    start_after = ["wait_before_stop_task"]
    target {
      key   = "Tasks"
      value = "app-ecs-tasks-${data.aws_default_tags.current.tags.environment-name}"
    }
  }

  action {
    action_id   = "aws:ecs:task-cpu-stress"
    description = null
    name        = "cpu_stress_100_percent"
    start_after = ["wait"]
    parameter {
      key   = "duration"
      value = "PT5M"
    }
    target {
      key   = "Tasks"
      value = "app-ecs-tasks-${data.aws_default_tags.current.tags.environment-name}"
    }
  }

  action {
    action_id   = "aws:ecs:task-network-latency"
    description = null
    name        = "ecs_network_latency"
    start_after = []
    parameter {
      key   = "duration"
      value = "PT5M"
    }
    target {
      key   = "Tasks"
      value = "app-ecs-tasks-${data.aws_default_tags.current.tags.environment-name}"
    }
  }

  action {
    action_id   = "aws:fis:wait"
    description = null
    name        = "wait"
    start_after = ["stop_task"]
    parameter {
      key   = "duration"
      value = "PT5M"
    }
  }

  action {
    action_id   = "aws:fis:wait"
    description = null
    name        = "wait_before_stop_task"
    start_after = ["ecs_network_latency"]
    parameter {
      key   = "duration"
      value = "PT5M"
    }
  }

  log_configuration {
    log_schema_version = 2

    cloudwatch_logs_configuration {
      log_group_arn = "${aws_cloudwatch_log_group.fis_app_ecs_tasks.arn}:*"
    }
  }

  stop_condition {
    source = "none"
    value  = null
  }

  target {
    name = "app-ecs-tasks-${data.aws_default_tags.current.tags.environment-name}"
    parameters = {
      "cluster" : var.ecs_cluster,
      "service" : aws_ecs_service.app.name,
    }
    resource_tag {
      key   = "aws:ecs:cluster-name"
      value = var.ecs_cluster
    }
    resource_type  = "aws:ecs:task"
    selection_mode = "ALL"
  }
}
