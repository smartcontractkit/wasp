# Tutorial
## Setup
Let's create our first load test
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
make dashboard
```
## Overview
General idea is to be able to compose load tests programmatically by combining different `Generators`

- `Generator` is an entity that can execute some workload using some `Gun` or `VU` definition, each `Generator` may have only one `Gun` or `VU` implementation used

- `Gun` can be an implementation of single or multiple sequential requests workload for stateless protocols

- `VU` is a stateful implementation that's more suitable for stateful protocols or when your client have some logic/simulating real users

- Each `Generator` have a `Schedule` that control workload params throughout the test (increase/decrease RPS or VUs)

- `Generators` can be combined to run multiple workload units in parallel or sequentially

- `Profiles` are wrappers that allow you to run multiple generators with different `Schedules` and wait for all of them to finish

- `AlertChecker` can be used in tests to check if any specific alerts with label and dashboardUUID was triggered and update test status

Load testing workflow can look like:
```mermaid
sequenceDiagram
    participant Product repo
    participant Runner
    participant K8s
    participant Loki
    participant Grafana
    participant Devs
    Product repo->>Product repo: Define NFR for different workloads<br/>Define application dashboard<br/>Define dashboard alerts<br/>Define load tests
    Product repo->>Grafana: Upload app dashboard<br/>Alerts has "requirement_name" label<br/>Each "requirement_name" groups is based on some NFR
    loop CI runs
    Product repo->>Runner: CI Runs small load test
    Runner->>Runner: Execute load test logic<br/>Run multiple generators
    Runner->>Loki: Stream load test data
    Runner->>Grafana: Checking "requirement_name": "baseline" alerts
    Grafana->>Devs: Notify devs (Dashboard URL/Alert groups)
    Product repo->>Runner: CI Runs huge load test
    Runner->>K8s: Split workload into multiple jobs<br/>Monitor jobs statuses
    K8s->>Loki: Stream load test data
    Runner->>Grafana: Checking "requirement_name": "stress" alerts
    Grafana->>Devs: Notify devs (Dashboard URL/Alert groups)
    end
```
## Examples

## RPS test
- [test](https://github.com/smartcontractkit/wasp/blob/master/examples/simple_rps/main.go#L9)
- [gun](https://github.com/smartcontractkit/wasp/blob/master/examples/simple_rps/gun.go#L23)
```
cd examples/simple_rps
go run .
```
Open [dashboard](http://localhost:3000/d/wasp/wasp-load-generator?orgId=1&var-test_group=generator_healthcheck&var-app=generator_healthcheck&var-cluster=generator_healthcheck&var-namespace=generator_healthcheck&var-branch=generator_healthcheck&var-commit=generator_healthcheck&from=now-5m&to=now&var-test_id=generator_healthcheck&var-gen_name=All&var-go_test_name=simple_rps&refresh=5s)

`Gun` must implement this [interface](https://github.com/smartcontractkit/wasp/blob/master/wasp.go#L36)

## VUs test
- [test](https://github.com/smartcontractkit/wasp/blob/master/examples/simple_instances/main.go#L10)
- [vu](https://github.com/smartcontractkit/wasp/blob/master/examples/simple_instances/instance.go#L34)
```
cd examples/simple_vu
go run .
```
Open [dashboard](http://localhost:3000/d/wasp/wasp-load-generator?orgId=1&var-test_group=generator_healthcheck&var-app=generator_healthcheck&var-cluster=generator_healthcheck&var-namespace=generator_healthcheck&var-branch=generator_healthcheck&var-commit=generator_healthcheck&from=now-5m&to=now&var-test_id=generator_healthcheck&var-gen_name=All&var-go_test_name=simple_instances&refresh=5s)

`VirtualUser` must implement this [interface](https://github.com/smartcontractkit/wasp/blob/master/wasp.go#L41)

## Profile test (group multiple generators in parallel)
- [test](https://github.com/smartcontractkit/wasp/blob/master/examples/profiles/main.go#L10)
- [gun](https://github.com/smartcontractkit/wasp/blob/master/examples/profiles/gun.go#L23)
- [vu](https://github.com/smartcontractkit/wasp/blob/master/examples/profiles/instance.go#L34)
```
cd examples/profiles
go run .
```
Open [dashboard](http://localhost:3000/d/wasp/wasp-load-generator?orgId=1&var-test_group=generator_healthcheck&var-app=generator_healthcheck&var-cluster=generator_healthcheck&var-namespace=generator_healthcheck&var-branch=generator_healthcheck&var-commit=generator_healthcheck&from=now-5m&to=now&var-test_id=generator_healthcheck&var-gen_name=All&var-go_test_name=my_test_ws&var-go_test_name=my_test&refresh=5s)

## Usage in tests
- [test](https://github.com/smartcontractkit/wasp/blob/master/examples/go_test/main_test.go#L15)
- [gun](https://github.com/smartcontractkit/wasp/blob/master/examples/go_test/gun.go#L23)
```
cd examples/go_test
go test -v -count 1 .
```
Open [dashboard](http://localhost:3000/d/wasp/wasp-load-generator?orgId=1&var-test_group=generator_healthcheck&var-app=generator_healthcheck&var-cluster=generator_healthcheck&var-namespace=generator_healthcheck&var-branch=generator_healthcheck&var-commit=generator_healthcheck&from=now-5m&to=now&var-test_id=generator_healthcheck&var-gen_name=All&var-go_test_name=TestProfile&refresh=5s)

## Scenario with simulating users behaviour
- [test](https://github.com/smartcontractkit/wasp/blob/master/examples/go_test/main_test.go#L15)
- [gun](https://github.com/smartcontractkit/wasp/blob/master/examples/go_test/gun.go#L23)
```
cd examples/scenario
go test -v -count 1 .
```
Open [dashboard](http://localhost:3000/d/wasp/wasp-load-generator?orgId=1&var-test_group=generator_healthcheck&var-app=generator_healthcheck&var-cluster=generator_healthcheck&var-namespace=generator_healthcheck&var-branch=generator_healthcheck&var-commit=generator_healthcheck&from=now-5m&to=now&var-test_id=generator_healthcheck&var-gen_name=All&var-go_test_name=TestProfile&refresh=5s)

## Defining NFRs and checking alerts
You can define different non-functional requirements groups
In this example we have 2 groups:
- `baseline` - checking both 99th latencies per `Generator` and errors
- `stress` - checking only errors

`WaspAlerts` can be defined on default `Generators` metrics, for each alert additional row is generated

`CustomAlerts` can be defined as [timeseries.Alert](https://pkg.go.dev/github.com/K-Phoen/grabana@v0.21.18/timeseries#Alert) but timeseries won't be included, though `AlertChecker` will check them

Run 2 tests, change mock latency/status codes to see how it works

Alert definitions usually defined with your `dashboard` and then constantly updated on each Git commit by your CI

After each run `AlertChecker` will fail the test if any alert from selected group was raised
- [definitions](https://github.com/smartcontractkit/wasp/blob/alerts_definitions/examples/alerts/main_test.go#L37)
- [wasp alerts](https://github.com/smartcontractkit/wasp/blob/alerts_definitions/examples/alerts/main_test.go#L40)
- [custom alerts](https://github.com/smartcontractkit/wasp/blob/alerts_definitions/examples/alerts/main_test.go#L82)
- [baseline NFR group test](https://github.com/smartcontractkit/wasp/blob/alerts_definitions/examples/alerts/main_test.go#L115)
- [stress NFR group test](https://github.com/smartcontractkit/wasp/blob/alerts_definitions/examples/alerts/main_test.go#L145)
```
cd examples/alerts
go test -v -count 1 -run TestBaselineRequirements
go test -v -count 1 -run TestStressRequirements
```
Open [alert groups](http://localhost:3000/alerting/groups)

Check [dashboard](http://localhost:3000/d/wasp/wasp-load-generator?orgId=1&refresh=5s&var-go_test_name=TestBaselineRequirement&var-go_test_name=TestBaselineRequirements&var-gen_name=All&var-branch=All&var-commit=All), you can see per alert timeseries in the bottom
