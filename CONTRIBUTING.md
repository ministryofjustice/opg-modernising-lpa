# Contributing

> We welcome contributions.

Please read this file to get a feel for what the expectations are.

- [Contributing](#contributing)
  - [Code of Conduct](#code-of-conduct)
  - [Coding Conventions](#coding-conventions)
  - [Opening pull requests](#opening-pull-requests)
  - [Commit messages](#commit-messages)

## Code of Conduct

Civil servants on this product all follow the [Civil Service Code](https://www.gov.uk/government/publications/civil-service-code/the-civil-service-code). External contributors should review the [Code of Conduct](CODE_OF_CONDUCT.md).

## Coding Conventions

For Go code we use errcheck, go-fmt-goimports and staticcheck.

For Terraform code, TFLint is used.

Code standards are enforced by the [pre-commit hooks](./.pre-commit-config.yaml) and the [build pipeline](./.github/workflows/). We recommend you install [pre-commit](https://pre-commit.com/) for local development.

## Opening pull requests

We have a pull request template, which will help you explain your work. It covers the purpose, approach and a checklist of key things to be sure of.

A passing PR build in Github Actions is required before a merge, along with approval from a member of the team.

We use a rebase workflow. Our primary branch is *main*. Please rebase branches on main if you need to pull in changes and use squash and merge for the final commit so we can back out changes.

## Commit messages

Explain what your work changes in the commit message and why it does so.
