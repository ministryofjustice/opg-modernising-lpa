# Configuring weblate access to manage translations and merge conflicts

These are instructions for configuring weblate access

## Prerequisites

install the weblate cli

```sh
pip install wlc
```

there are instructions also for using the docker image here: `https://docs.weblate.org/en/weblate-5.2.1/wlc.html#docker-usage`

## Configuring weblate access

Put the following in your `~/.config/weblate` file:

```ini
[weblate]
url = https://moj.weblate.cloud/api/

[keys]
https://moj.weblate.cloud/api/ = APIKEY
```

Your api key can be found in the weblate account page under `https://moj.weblate.cloud/accounts/profile/#api`

## Managing translations

The repository has a `.weblate` file which contains the configuration for our translations.

That means from the repository root you can run weblate cli commands like:

```sh
wlc ls
```

or

```sh
wlc lock
```
