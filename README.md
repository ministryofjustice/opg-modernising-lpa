# OPG Modernising LPA

![path-to-live-workflow](https://github.com/ministryofjustice/opg-modernising-lpa/actions/workflows/workflow_path_to_live.yml/badge.svg)
![licence-mit](https://img.shields.io/github/license/ministryofjustice/opg-modernising-lpa-docs.svg)
[![codecov](https://codecov.io/gh/ministryofjustice/opg-modernising-lpa/branch/main/graph/badge.svg?token=mKTlhc906S)](https://codecov.io/gh/ministryofjustice/opg-modernising-lpa)

[![repo standards badge](https://img.shields.io/badge/dynamic/json?color=blue&style=for-the-badge&logo=github&label=MoJ%20Compliant&query=%24.result&url=https%3A%2F%2Foperations-engineering-reports.cloud-platform.service.justice.gov.uk%2Fapi%2Fv1%2Fcompliant_public_repositories%2Fopg-modernising-lpa)](https://operations-engineering-reports.cloud-platform.service.justice.gov.uk/public-github-repositories.html#opg-modernising-lpa "Link to report")

## OPG Modernising LPA Documentation

Documentation for the service can be found [in the /docs/ folder](./docs/README.md).

## Getting Started

### Prerequisites

* Docker and docker-compose
* Nodejs and Yarn

### Installation

Install dependencies for development

```shell
yarn install
```

Bring the service up

```shell
docker compose up -d
```

### Run Cypress tests

```shell
make run-cypress
```

### Local development

To run the app in dev mode on arm64/apple silicon:

```shell
make app-up-build-dev
```

or the following for amd64/intel:

```shell
make app-up-build-dev-amd
```

Dev mode adds hot reloading via [air](https://github.com/cosmtrek/air) which will watch `.go` and `.gohtml` for changes and recompile the app.

It also enables debugging via [delve](https://github.com/go-delve/delve). Delve runs on `localhost:2345` - add this to your editor/IDE debug config settings and then start adding breakpoints to step through the app code when running on localhost. Example VSCode confige:

```json
{
    "name": "Go remote debug",
    "type": "go",
    "request": "attach",
    "mode": "remote",
    "port": 2345,
    "host": "127.0.0.1",
    "trace": "verbose",
}
```

## Licence

The OPG Modernising LPA code in this repository is released under the MIT license, a copy of which can be found in [LICENCE](./LICENCE).
