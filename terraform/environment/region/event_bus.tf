module "event_bus" {
  source               = "./modules/event_bus"
  target_event_bus_arn = var.target_event_bus_arn
  iam_role             = var.iam_roles.cross_account_put
  receive_account_ids  = var.receive_account_ids
  providers = {
    aws.region = aws.region
  }
}
