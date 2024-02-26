<p align="center">
    <img alt="wasp" src="./docs/wasp-2.png"> 
</p>
<div align="center">

![GitHub release (latest SemVer)](https://img.shields.io/github/v/release/smartcontractkit/wasp)
![GitHub](https://img.shields.io/github/license/smartcontractkit/wasp)
[![Go Report Card](https://goreportcard.com/badge/github.com/smartcontractkit/wasp)](https://goreportcard.com/report/github.com/smartcontractkit/wasp)
[![Go Tests](https://github.com/smartcontractkit/wasp/actions/workflows/test.yml/badge.svg)](https://github.com/smartcontractkit/wasp/actions/workflows/test.yml)
[![Bench](https://github.com/smartcontractkit/wasp/actions/workflows/bench.yml/badge.svg?branch=master)](https://github.com/smartcontractkit/wasp/actions/workflows/bench.yml)
<a href='https://github.com/jpoles1/gopherbadger' target='_blank'>![gopherbadger-tag-do-not-edit](https://img.shields.io/badge/Go%20Coverage-80%25-brightgreen.svg?longCache=true&style=flat)</a>

Scalable protocol-agnostic load testing library for `Go`

</div>

## Goals
- Easy to reuse any custom client `Go` code
- Easy to grasp
- Have a slim codebase (500-1k loc)
- No test harness or CLI, easy to integrate and run with plain `go test`
- Have a predictable performance footprint
- Easy to create synthetic or user-based scenarios
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
![dashboard_img](./docs/dashboard_basic.png)

Remove environment:
```bash
make stop
```

## Test Layout and examples
Check [examples](examples/README.md) to understand what is the easiest way to structure your tests, run them both locally and remotely, at scale, inside `k8s`

## Run pyroscope test
```
make pyro_start
make test_pyro_rps
make test_pyro_vu
make pyro_stop
```
Open [pyroscope](http://localhost:4040/)

You can also use `trace.out` in the root folder with `Go` default tracing UI

## How it works
![img.png](docs/how-it-works.png)

Check this [doc](./HOW_IT_WORKS.md) for more examples and project overview

## Loki debug
You can check all the messages the tool sends with env var `WASP_LOG_LEVEL=trace`

If Loki client fail to deliver a batch test will proceed, if you experience Loki issues, consider setting `Timeout` in `LokiConfig` or set `MaxErrors: 10` to return an error after N Loki errors

`MaxErrors: -1` can be used to ignore all the errors

Default Promtail settings are:
```
&LokiConfig{
    TenantID:                os.Getenv("LOKI_TENANT_ID"),
    URL:                     os.Getenv("LOKI_URL"),
    Token:                   os.Getenv("LOKI_TOKEN"),
    BasicAuth:               os.Getenv("LOKI_BASIC_AUTH"),
    MaxErrors:               10,
    BatchWait:               5 * time.Second,
    BatchSize:               500 * 1024,
    Timeout:                 20 * time.Second,
    DropRateLimitedBatches:  false,
    ExposePrometheusMetrics: false,
    MaxStreams:              600,
    MaxLineSize:             999999,
    MaxLineSizeTruncate:     false,
}
```
If you see errors like
```
ERR Malformed promtail log message, skipping Line=["level",{},"component","client","host","...","msg","batch add err","tenant","","error",{}]
```
Try to increase `MaxStreams` even more or check your `Loki` configuration
