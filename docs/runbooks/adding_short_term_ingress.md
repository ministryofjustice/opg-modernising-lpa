# Adding short term ingress

## Overview

We sometimes need to temporarily allow access to a service from a specific IP address or range of IP addresses. Instead of making changes to the allow-list repository, we maintain a short-term ingress list as a parameter store in AWS Systems Manager.

## Adding an IP address to short term ingress for an account

1. Sign in to the AWS Management Console and assume the operator role into the Management account, in the us-east-1 region.
1. Navigate to the AWS Systems Manager, and then to the Parameter Store.
1. Search for the parameter `/modernising-lpa/additional-allowed-ingress-cidrs/<account-name>` and click on it.
1. Click on the `Edit` button.
1. Add the IP address or range of IP addresses to the `Value` field as comma-separated values. IP addresses should be in CIDR notation. for example a single IP address would be `123.456.789.0/32` and a range of IP addresses would be `123.456.789.0/24`.
1. Click on the `Save changes` button.
1. Lastly, a deployment of the environment is required to apply the changes.

Remember to remove the IP address or range of IP addresses from the short-term ingress list once they are no longer required.
