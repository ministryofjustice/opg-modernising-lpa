# Restoring a DynamoDB Backup

AWS Backup is used to schedule and manage backups for each dynamodb table.
As with other AWS backup processes, restored backups create a new resource which must be named differently from the resource. If the source were deleted prior to restoration this would not be an
issue, but this comes with risk.
This document will walk you through how to restore a backup, and bring it into service, and how to delete the old table.
The guide is targetted at Production restoring a backup of the ActorUsers table, but the screenshots are from Development.

## Before Proceeding

- This procedure requires pairing. This is to help validate each step is completed correctly and have more of the team comfortable with the procedure.
- It is strongly recommended that the service is put into maintenance mode to prevent users entering data that will be lost. Customer data entered since the last backup will also be lost.
- We will need to initiate a Change Freeze when we are ready to bring the restored tables into the service. During this time, any run of the path to live pipeline risks deleting the restored or outgoing
tables. The freeze will help to prevent this.
- This restore procedure can take hours to perform.
- When running this procedure against Production or Preproduction, you will need to be able to assume the breakglass role. If you cannot, you do not have the required permission set to perform these
tasks.
- You will need the image tag currently deployed to production

## Restore a table from a backup

1. Sign in to the AWS Console, Assume the breakglass role in the Production account, and navigate to AWS Backup.

1. From the menu on the left, expand My account and click on Backup Vaults.

1. Click on the vault named `eu-west-1-production-backup-vault`. This will show a list of backups for each table that can be used as recovery points.

1. Select a single backup using the Resource ID and Creation time to pick one that is appropriate, and tick it. At the top right of this table, click the Actionsdropdown and choose Restore.

1. This will open the Restore backup wizard. Choose a new name for the table. Use the original name plus a `-` then the date of restoration in the format `YYYYMMDD`. For example `production-Lpas-20240715`. This will make is
easier to manage restored tables going forward. It is not possible to change the name of a DynamoDB after it is created. This new name will be brought into our infrastructure as code. Note that indexes will also be restored.

1. You’ll be taken to the Jobs page on the Restore jobs tab. Restore jobs can take a long time (hours) to complete.

## Bring restored table into service

1. Here we will update the infrastructure as code to use the new restored table. Ensure you are up to date with the main branch, and create a new branch.

1. Edit the terraform/environment/.envrc file to set the `TF_VAR_default_role` to `breakglass`. Run `direnv allow` to apply the changes.

1. In terminal, navigate to the Terraform environment configuration and select the production workspace. This will require breakglass permissions

    ```bash
    cd terraform/environments
    aws-vault exec identity -- terraform init
    aws-vault exec identity -- terraform workspace select production
    ```

1. Remove the dynamodb table from the terraform configuration

    ```shell
    aws-vault exec identity -- terraform state rm aws_dynamodb_table.lpas_table
    ```

1. Next import the restored table using the new name

    ```shell
    aws-vault exec identity -- terraform import aws_dynamodb_table.lpas_table arn:aws:dynamodb:eu-west-1:<prod-account-id>:table/production-Lpas-20240715
    ```

1. Next, update the name of the new table in terraform.tfvars.json for the environment, for example

    ```json
    {
            },
      "dynamodb": {
        "table_name": "Lpas-20240715",
        "region_replica_enabled": true
      },
    }
    ```

1. From there we can run a plan to check what will happen.

    ```shell
    aws-vault exec identity -- terraform plan
    ```

    We are expecting to see updates to our restored dynamoDB table, and changes to services and resources that reference the table name or ARN.

    > Things to check for
    > Policy Documents for API and Admin updating to use new (restored table)
    > AWS Backup managing the new table
    > DynamoDB Table tags, point in time restore enabled, server side encryption enabled and TTL activation
    > ECS Services and Task Definition updates for API and Admin
    > Plans and Applies always produce a Config file.

1. Once happy with the plan, apply the changes

    ```shell
    aws-vault exec identity -- terraform apply
    ```

1. Commit our changes to the DynamoDB table names, and raise a PR to ensure these persist.
    Once this PR is merged and has reached production, we can release the change freeze.


## Delete the old table

1. At this point we can delete the old tables. They are no longer managed by Terraform, so we must do this in the AWS console.
    In the AWS console, again while assuming the breakglass role in the production account, navigate to the DynamoDB console.
    Select Tables from the menu on the left.

Select the tables that we want to delete by ticking them and click on Delete at the top right of the Tables table.

On the delete dialogue, choose the option to delete all Cloudwatch alarms for the table(s), follow the prompt to confirm that you want the table(s), and click Delete table.

Once deletion of the tables no longer required is completed, so too is the DynamoDB Table restore procedure.
