module "reduced_fees" {
  count                = local.environment.reduced_fees.enabled ? 1 : 0
  source               = "./reduced_fees"
  target_event_bus_arn = local.environment.reduced_fees.target_event_bus_arn
  providers = {
    aws.region = aws.eu_west_1
    aws.global = aws.global
  }
}

moved {
  from = module.reduced_fees
  to   = module.reduced_fees[0]
}
