module "event_bus" {
  source               = "./modules/event_bus"
  target_event_bus_arn = var.target_event_bus_arn
  providers = {
    aws.region = aws.region
  }
}
