output "workspace_name" {
  value = terraform.workspace
}

variable "pagerduty_api_key" {
  type      = string
  sensitive = true
}

variable "accounts" {
  type = map(
    object({
      account_id             = string
      account_name           = string
      is_production          = bool
      regions                = list(string)
      pagerduty_service_name = string
    })
  )
}

locals {
  account_name = lower(replace(terraform.workspace, "_", "-"))
  account      = var.accounts[local.account_name]

  mandatory_moj_tags = {
    business-unit    = "OPG"
    application      = "opg-modernising-lpa"
    environment-name = local.account.account_name
    owner            = "OPG Webops: opgteam+modernising-lpa@digital.justice.gov.uk"
    is-production    = local.account.is_production
    runbook          = "https://github.com/ministryofjustice/opg-modernising-lpa"
    source-code      = "https://github.com/ministryofjustice/opg-modernising-lpa"
  }


  optional_tags = {
    account-name           = local.account.account_name
    infrastructure-support = "OPG Webops: opgteam+modernising-lpa@digital.justice.gov.uk"
  }

  default_tags = merge(local.mandatory_moj_tags, local.optional_tags)
}
