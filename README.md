# Wasp

![GitHub release (latest SemVer)](https://img.shields.io/github/v/release/smartcontractkit/wasp)
[![Go Report Card](https://goreportcard.com/badge/github.com/smartcontractkit/wasp)](https://goreportcard.com/report/github.com/smartcontractkit/wasp)
[![Go Tests](https://github.com/smartcontractkit/wasp/actions/workflows/test.yml/badge.svg)](https://github.com/smartcontractkit/wasp/actions/workflows/test.yml)
<a href='https://github.com/jpoles1/gopherbadger' target='_blank'>![gopherbadger-tag-do-not-edit](https://img.shields.io/badge/Go%20Coverage-80%25-brightgreen.svg?longCache=true&style=flat)</a>

A simple protocol-agnostic load testing tool for `Go`

## Goals
- Easy to reuse any custom client `Go` code
- Easy to grasp
- Have slim codebase (500-1k loc)
- Have predictable performance footprint when tested with protocol mocks
- Be able to perform synthetic load testing for request-based protocols in `Go` with `RPS bound load` (http, etc.)
- Be able to perform VU load testing for streaming protocols/scenario tests in `Go` with `VU bound load` (ws, scenario, etc.)
- Scalable in `k8s` without complicated configuration or vendored UI interfaces
- Non-opinionated reporting, push any data to `Loki`
## Setup
We are using `nix` for deps, see [installation](https://nixos.org/manual/nix/stable/installation/installation.html) guide
```bash
nix develop
```

## Run example tests with Grafana + Loki
```bash
make start
```
Insert `GRAFANA_TOKEN` created in previous command
```bash
export LOKI_URL=http://localhost:3030/loki/api/v1/push
export GRAFANA_URL=http://localhost:3000
export GRAFANA_TOKEN=...
export DATA_SOURCE_NAME=Loki
export DASHBOARD_FOLDER=LoadTests
export WASP_LOG_LEVEL=info
make dashboard
```
Run some tests:
```
make test_loki
```
Open your [Grafana dashboard](http://localhost:3000/d/wasp/wasp-load-generator?orgId=1&refresh=5s)

Basic [dashboard](dashboard/dashboard.go):
![dashboard_img](docs/dashboard_basic.png)

Remove environment:
```bash
make stop
```

## Run pyroscope test
```
make pyro_start
make test_pyro
make pyro_stop
```
Open [pyroscope](http://localhost:4040/)

You can also use `trace.out` in the root folder with `Go` default tracing UI

## Tutorial
Check [tutorial](./TUTORIAL.md) for more examples

## Loki debug
You can check all the messages the tool sends with env var `WASP_LOG_LEVEL=trace`