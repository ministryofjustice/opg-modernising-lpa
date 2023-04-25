# Manage Maintenance Mode

This script will enable or disable maintenance mode for a targeted environment.

## Usage

This script can be run from github actions or a local command line with the correct permissions.

### Github Actions

Navigate to [Github Actions](https://github.com/ministryofjustice/opg-modernising-lpa/actions) for the MLPAB service

Click on the `[WD] Manage Maintenance Mode` workflow

Click on the `Run workflow` button

Enter the `environment` and `region` you wish to run the manage maintenance script against, and select the `maintenance mode enabled` option `true` or `false`.

Click on the `Run workflow` button

### Local Command Line

To turn on maintenance mode from the command line, run the following command:

``` bash
aws-vault exec mod-lpa-prod -- ./scripts/manage_maintenance.sh \
  --environment production \
  --maintenance_mode
```

To turn off maintenance mode from the command line, run the following command:

``` bash
aws-vault exec mod-lpa-prod -- ./scripts/manage_maintenance.sh \
  --environment production \
  --disable_maintenance_mode
```
