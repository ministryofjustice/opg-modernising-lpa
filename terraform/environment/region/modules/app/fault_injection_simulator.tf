resource "aws_fis_experiment_template" "ecs_app" {
  provider    = aws.region
  description = "${data.aws_default_tags.current.tags.environment-name} - APP ECS Task Experiments"
  role_arn    = var.fault_injection_simulator_role_arn

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
