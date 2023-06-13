module "global" {
  source = "./global"
  providers = {
    aws.global = aws.global
  }
}
