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

## adding weblate remote

You can add a weblate remote to your local repository.

```sh
git remote add weblate https://moj.weblate.cloud/git/opg-modernising-lpa/opg-modernising-lpa/
git remote update weblate
```

to set the remote back to origin run:

```sh
git remote update origin
```

## resolving merge conflicts

The way to resolve merge conflicts is to retrieve the changes from weblate onto a new branch and merge them. After they're on main, we can either refresh or reset the project in weblate.

Commit all pending changes in Weblate and lock the translation component.

```sh
wlc commit; wlc lock
```

Create a new branch for Weblate changes (name is arbitrary).

```sh
git checkout -b weblate-resolve-merge-conflicts
```

Switch to the weblate remote, and merge Weblate changes and resolve any conflicts.

```sh
git remote update weblate
git merge weblate/main
```

set the remote back to origin, rebase Weblate changes on top of main and resolve any conflicts.

```sh
git remote update origin
git rebase main
```

Push changes into upstream repository.

```sh
git push weblate-resolve-merge-conflicts
```

When the PR is merged, Weblate should now be able to see updated repository and you can unlock it.

```sh
wlc pull ; wlc unlock
```
