module "fault_injection_simulator_experiments" {
  count  = var.fault_injection_enabled ? 1 : 0
  source = "./modules/fault_injection_simulator_experiments"
  providers = {
    aws.region = aws.region
  }
  fault_injection_simulator_role = var.iam_roles.fault_injection_simulator

}
