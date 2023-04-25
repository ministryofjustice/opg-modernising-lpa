# Manage Maintenance Mode

This script will enable or disable maintenance mode for a targeted environment.

## Usage

This script can be run from github actions or a local command line with the correct permissions.

### Github Actions




### Local Command Line

To turn on maintenance mode for both the use and view front ends

``` bash
aws-vault exec mod-lpa-prod -- ./scripts/manage_maintenance.sh \
  --environment production \
  --maintenance_mode
```

To turn off maintenance mode for both the use and view front ends

``` bash
aws-vault exec mod-lpa-prod -- ./scripts/manage_maintenance.sh \
  --environment production \
  --disable_maintenance_mode
```
