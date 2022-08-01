# Terraform Shared

This terraform configuration manages per-environment resources.

Per-account or otherwise shared resources are managed in `../account`

## Namespace resources

It is important to namespace resources to avoid getting errors for creating resources that already exist.

There are two namespace variables available.

```hcl
"${local.environment_name}"
```

can return `uml93` or `production`


## Running Terraform Locally

This repository comes with an `.envrc` file containing useful environment variables for working with this repository.

`.envrc` can be sourced automatically using either [direnv](https://direnv.net) or manually with bash.

```shell
source .envrc
```

```shell
direnv allow
```

This sets environment variables that allow the following commands with no further setup

```shell
aws-vault exec identity -- terraform init
aws-vault exec identity -- terraform plan
aws-vault exec identity -- terraform force-unlock 49b3784c-51eb-668d-ac4b-3bd5b8701925
```

## Fixing state lock issue

A Terraform state lock error can happen if a terraform job is forcefully terminated (normal ctrl+c gracefully releases state lock).

CircleCI terminates a process if you cancel a job, so state lock doesn't get released.

Here's how to fix it if it happens.
Error:

```shell
rror locking state: Error acquiring the state lock: ConditionalCheckFailedException: The conditional request failed
    status code: 400, request id: 60Q304F4TMIRB13AMS36M49ND7VV4KQNSO5AEMVJF66Q9ASUAAJG
Lock Info:
  ID:        69592de7-6132-c863-ae53-976776ffe6cf
  Path:      opg.terraform.state/env:/development/opg-modernising-lpa/terraform.tfstate
  Operation: OperationTypeApply
  Who:       @d701fcddc381
  Version:   0.11.13
  Created:   2019-05-09 16:01:50.027392879 +0000 UTC
  Info:
```

Fix:

```shell
aws-vault exec identity -- terraform init
aws-vault exec identity -- terraform workspace select development
aws-vault exec identity -- terraform force-unlock 69592de7-6132-c863-ae53-976776ffe6cf
```

It is important to select the correct workspace.
For terraform_environment, this will be based on your PR and can be found in the Github Actions pipeline job `PR Environment Deploy`
