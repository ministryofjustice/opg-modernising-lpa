resource "aws_appautoscaling_target" "ecs_service" {
  provider           = aws.region
  service_namespace  = "ecs"
  resource_id        = "service/${var.aws_ecs_cluster_name}/${var.aws_ecs_service_name}"
  scalable_dimension = "ecs:service:DesiredCount"
  role_arn           = var.ecs_autoscaling_service_role_arn
  max_capacity       = var.ecs_task_autoscaling_maximum
  min_capacity       = var.ecs_task_autoscaling_minimum
}

# Automatically scale capacity up by one
resource "aws_appautoscaling_policy" "up" {
  provider           = aws.region
  name               = "${var.environment_name}-${var.aws_ecs_service_name}-scale-up"
  service_namespace  = "ecs"
  resource_id        = aws_appautoscaling_target.ecs_service.resource_id
  scalable_dimension = aws_appautoscaling_target.ecs_service.scalable_dimension

  step_scaling_policy_configuration {
    adjustment_type         = "ChangeInCapacity"
    cooldown                = var.scale_up_cooldown
    metric_aggregation_type = "Maximum"

    step_adjustment {
      metric_interval_lower_bound = 0
      scaling_adjustment          = 1
    }
  }

  depends_on = [aws_appautoscaling_target.ecs_service]
}

# Automatically scale capacity down by one
resource "aws_appautoscaling_policy" "down" {
  provider           = aws.region
  name               = "${var.environment_name}-${var.aws_ecs_service_name}-scale-down"
  service_namespace  = "ecs"
  resource_id        = aws_appautoscaling_target.ecs_service.resource_id
  scalable_dimension = aws_appautoscaling_target.ecs_service.scalable_dimension

  step_scaling_policy_configuration {
    adjustment_type         = "ChangeInCapacity"
    cooldown                = var.scale_down_cooldown
    metric_aggregation_type = "Maximum"

    step_adjustment {
      metric_interval_lower_bound = 0
      scaling_adjustment          = -1
    }
  }

  depends_on = [aws_appautoscaling_target.ecs_service]
}

resource "aws_cloudwatch_metric_alarm" "scale_up" {
  provider                  = aws.region
  alarm_name                = "${var.environment_name}-${var.region_name}-${var.aws_ecs_service_name}-scale-up"
  comparison_operator       = "GreaterThanOrEqualToThreshold"
  evaluation_periods        = "2"
  threshold                 = "1"
  alarm_description         = "Scale up based on Mem, Cpu and Task Count"
  insufficient_data_actions = []

  metric_query {
    id          = "up"
    expression  = "IF((cpu > ${var.autoscaling_metric_max_cpu_target} OR mem > ${var.autoscaling_metric_max_memory_target}) AND tc < ${var.ecs_task_autoscaling_maximum}, 1, 0)"
    label       = "ContainerScaleUp"
    return_data = "true"
  }

  metric_query {
    id = "cpu"

    metric {
      metric_name = "CPUUtilization"
      namespace   = "AWS/ECS"
      period      = "60"
      stat        = "Maximum"

      dimensions = {
        ServiceName = var.aws_ecs_service_name
        ClusterName = var.aws_ecs_cluster_name
      }
    }
  }

  metric_query {
    id = "mem"

    metric {
      metric_name = "MemoryUtilization"
      namespace   = "AWS/ECS"
      period      = "60"
      stat        = "Average"

      dimensions = {
        ServiceName = var.aws_ecs_service_name
        ClusterName = var.aws_ecs_cluster_name
      }
    }
  }

  metric_query {
    id = "tc"

    metric {
      metric_name = "DesiredTaskCount"
      namespace   = "ECS/ContainerInsights"
      period      = "60"
      stat        = "Average"
      unit        = "Count"

      dimensions = {
        ServiceName = var.aws_ecs_service_name
        ClusterName = var.aws_ecs_cluster_name
      }
    }
  }
  alarm_actions = [aws_appautoscaling_policy.up.arn]
}

resource "aws_cloudwatch_metric_alarm" "scale_down" {
  provider                  = aws.region
  alarm_name                = "${var.environment_name}-${var.region_name}-${var.aws_ecs_service_name}-scale-down"
  comparison_operator       = "GreaterThanOrEqualToThreshold"
  evaluation_periods        = "2"
  threshold                 = "1"
  alarm_description         = "Scale down based on Mem, Cpu and Task Count"
  insufficient_data_actions = []

  metric_query {
    id          = "down"
    expression  = "IF((cpu < ${var.autoscaling_metric_min_cpu_target} AND mem < ${var.autoscaling_metric_min_memory_target}) AND tc > ${var.ecs_task_autoscaling_minimum}, 1, 0)"
    label       = "ContainerScaleDown"
    return_data = "true"
  }

  metric_query {
    id = "cpu"

    metric {
      metric_name = "CPUUtilization"
      namespace   = "AWS/ECS"
      period      = "60"
      stat        = "Average"

      dimensions = {
        ServiceName = var.aws_ecs_service_name
        ClusterName = var.aws_ecs_cluster_name
      }
    }
  }

  metric_query {
    id = "mem"

    metric {
      metric_name = "MemoryUtilization"
      namespace   = "AWS/ECS"
      period      = "60"
      stat        = "Average"

      dimensions = {
        ServiceName = var.aws_ecs_service_name
        ClusterName = var.aws_ecs_cluster_name
      }
    }
  }

  metric_query {
    id = "tc"

    metric {
      metric_name = "DesiredTaskCount"
      namespace   = "ECS/ContainerInsights"
      period      = "60"
      stat        = "Average"
      unit        = "Count"

      dimensions = {
        ServiceName = var.aws_ecs_service_name
        ClusterName = var.aws_ecs_cluster_name
      }
    }
  }
  alarm_actions = [aws_appautoscaling_policy.down.arn]
}

resource "aws_cloudwatch_metric_alarm" "max_scaling_reached" {
  provider                  = aws.region
  alarm_name                = "${var.environment_name}-${var.region_name}-${var.aws_ecs_service_name}-max-scaling-reached"
  alarm_actions             = var.max_scaling_alarm_actions
  comparison_operator       = "GreaterThanOrEqualToThreshold"
  evaluation_periods        = "2"
  metric_name               = "RunningTaskCount"
  namespace                 = "ECS/ContainerInsights"
  period                    = "30"
  statistic                 = "Average"
  threshold                 = var.ecs_task_autoscaling_maximum
  alarm_description         = "This metric monitors ecs running task count for the ${var.environment_name}-${var.region_name} ${var.aws_ecs_service_name} service"
  insufficient_data_actions = []
  dimensions = {
    ServiceName = var.aws_ecs_service_name
    ClusterName = var.aws_ecs_cluster_name
  }
}
