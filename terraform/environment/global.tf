module "global" {
  source                                  = "./global"
  cloudwatch_application_insights_enabled = local.environment.app.cloudwatch_application_insights_enabled
  providers = {
    aws.global = aws.global
  }
}
