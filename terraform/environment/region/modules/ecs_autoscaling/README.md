# ECS Autoscaling

This module creates autoscaling policies for ECS services that track a combined CPU and memory metric.

## Providers

| Name | Version |
|------|---------|
| aws  | 3.38.0  |

## Resources

| Name | Type |
|------|------|
| [aws_appautoscaling_policy.down](https://registry.terraform.io/providers/hashicorp/aws/latest/docs/resources/appautoscaling_policy) | resource |
| [aws_appautoscaling_policy.up](https://registry.terraform.io/providers/hashicorp/aws/latest/docs/resources/appautoscaling_policy) | resource |
| [aws_appautoscaling_target.target](https://registry.terraform.io/providers/hashicorp/aws/latest/docs/resources/appautoscaling_target) | resource |
| [aws_cloudwatch_metric_alarm.max_scaling_reached](https://registry.terraform.io/providers/hashicorp/aws/latest/docs/resources/cloudwatch_metric_alarm) | resource |
| [aws_cloudwatch_metric_alarm.scale_down](https://registry.terraform.io/providers/hashicorp/aws/latest/docs/resources/cloudwatch_metric_alarm) | resource |
| [aws_cloudwatch_metric_alarm.scale_up](https://registry.terraform.io/providers/hashicorp/aws/latest/docs/resources/cloudwatch_metric_alarm) | resource |

## Inputs

| Name | Description | Type | Default | Required |
|------|-------------|------|---------|:--------:|
| <a name="input_autoscaling_metric_max_cpu_target"></a> [autoscaling\_metric\_max\_cpu\_target](#input\_autoscaling\_metric\_max\_cpu\_target) | The target value for the CPU metric. | `number` | `80` | no |
| <a name="input_autoscaling_metric_max_memory_target"></a> [autoscaling\_metric\_max\_memory\_target](#input\_autoscaling\_metric\_max\_memory\_target) | The target value for the memory metric. | `number` | `80` | no |
| <a name="input_autoscaling_metric_min_cpu_target"></a> [autoscaling\_metric\_min\_cpu\_target](#input\_autoscaling\_metric\_min\_cpu\_target) | The target value for the CPU metric. | `number` | `30` | no |
| <a name="input_autoscaling_metric_min_memory_target"></a> [autoscaling\_metric\_min\_memory\_target](#input\_autoscaling\_metric\_min\_memory\_target) | The target value for the memory metric. | `number` | `30` | no |
| <a name="input_aws_ecs_cluster_name"></a> [aws\_ecs\_cluster\_name](#input\_aws\_ecs\_cluster\_name) | Name of the ECS cluster for the service being scaled. | `string` | n/a | yes |
| <a name="input_aws_ecs_service_name"></a> [aws\_ecs\_service\_name](#input\_aws\_ecs\_service\_name) | Name of the ECS service. | `string` | n/a | yes |
| <a name="input_ecs_autoscaling_service_role_arn"></a> [ecs\_autoscaling\_service\_role\_arn](#input\_ecs\_autoscaling\_service\_role\_arn) | The ARN of the IAM role that allows Application AutoScaling to modify your scalable target on your behalf. | `string` | n/a | yes |
| <a name="input_ecs_task_autoscaling_maximum"></a> [ecs\_task\_autoscaling\_maximum](#input\_ecs\_task\_autoscaling\_maximum) | The max capacity of the scalable target. | `number` | n/a | yes |
| <a name="input_ecs_task_autoscaling_minimum"></a> [ecs\_task\_autoscaling\_minimum](#input\_ecs\_task\_autoscaling\_minimum) | The min capacity of the scalable target. | `number` | `1` | no |
| <a name="input_environment"></a> [environment](#input\_environment) | Name of the environment. | `string` | n/a | yes |
| <a name="input_max_scaling_alarm_actions"></a> [max\_scaling\_alarm\_actions](#input\_max\_scaling\_alarm\_actions) | List of alarm actions for maximum autoscaling being reached. | `list(string)` | n/a | yes |
| <a name="input_scale_down_cooldown"></a> [scale\_down\_cooldown](#input\_scale\_down\_cooldown) | The amount of time, in seconds, after a scale in activity completes before another scale in activity can start. | `number` | `60` | no |
| <a name="input_scale_up_cooldown"></a> [scale\_up\_cooldown](#input\_scale\_up\_cooldown) | The amount of time, in seconds, after a scale out activity completes before another scale out activity can start. | `number` | `60` | no |

## Outputs

No output.
