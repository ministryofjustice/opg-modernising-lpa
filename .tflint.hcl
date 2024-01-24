config {
  format = "compact"
}

plugin "terraform" {
  enabled = true
  preset  = "recommended"
}

rule "terraform_required_providers" {
  enabled = true

  # defaults
  source = true
  version = false
}
