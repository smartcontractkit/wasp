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

- `ClusterProfiles` are high-level wrappers that create multiple profile parts and scale your test in `k8s`

- `VU` implementations can also include sequential and parallel requests to simulate users behaviour

- `AlertChecker` can be used in tests to check if any specific alerts with label and dashboardUUID was triggered and update test status

Example cluster execution diagram:
```mermaid
---
title: Workload execution. P - Profile, G - Generator, VU - VirtualUser
---
flowchart TB
    ClusterProfile-- generate k8s manifests/deploy/await jobs completion -->P1
    ClusterProfile-->PN
    ClusterProfile-- check NFRs -->Grafana
    subgraph Pod1
    P1-->P1-G1
    P1-->P1-GN
    P1-G1-->P1-G1-VU1
    P1-G1-->P1-G1-VUN
    P1-GN-->P1-GN-VU1
    P1-GN--->P1-GN-VUN

    P1-G1-VU1-->P1-Batch
    P1-G1-VUN-->P1-Batch
    P1-GN-VU1-->P1-Batch
    P1-GN-VUN-->P1-Batch
    end
    subgraph PodN
    PN-->PN-G1
    PN-->PN-GN
    PN-G1-->PN-G1-VU1
    PN-G1-->PN-G1-VUN
    PN-GN-->PN-GN-VU1
    PN-GN--->PN-GN-VUN

    PN-G1-VU1-->PN-Batch
    PN-G1-VUN-->PN-Batch
    PN-GN-VU1-->PN-Batch
    PN-GN-VUN-->PN-Batch

    end
    P1-Batch-->Loki
    PN-Batch-->Loki

    Loki-->Grafana


```

For now, only `one node` mode is available, `k8s` scaling is planned.

Example `Syntetic/RPS` test diagram:

```mermaid
---
title: Syntetic/RPS test
---
sequenceDiagram
    participant Profile(Test)
    participant Scheduler
    participant Generator(Gun)
    participant Promtail
    participant Loki
    participant Grafana
    loop Test Execution
        Profile(Test) ->> Generator(Gun): Start with (API, TestData)
        loop Schedule
            Scheduler ->> Scheduler: Process schedule segment
            Scheduler ->> Generator(Gun): Set new RPS target
            loop RPS load
                Generator(Gun) ->> Generator(Gun): Execute Call() in parallel
                Generator(Gun) ->> Promtail: Save CallResult
            end
            Promtail ->> Loki: Send batch<br/>when ready or timeout
        end
        Scheduler ->> Scheduler: All segments done<br/>wait all responses<br/>test ends
        Profile(Test) ->> Grafana: Check alert groups<br/>FAIL or PASS the test
    end
```

Example `VUs` test diagram:

```mermaid
---
title: VUs test
---
sequenceDiagram
    participant Profile(Test)
    participant Scheduler
    participant Generator(VUs)
    participant VU1
    participant VU2
    participant Promtail
    participant Loki
    participant Grafana
    loop Test Execution
        Profile(Test) ->> Generator(VUs): Start with (API, TestData)
        loop Schedule
            Scheduler ->> Scheduler: Process schedule segment
            Scheduler ->> Generator(VUs): Set new VUs target
            loop VUs load
                Generator(VUs) ->> Generator(VUs): Add/remove VUs
                Generator(VUs) ->> VU1: Start/end
                Generator(VUs) ->> VU2: Start/end
                VU1 ->> VU1: Run loop, execute multiple calls
                VU1 ->> Promtail: Save []CallResult
                VU2 ->> VU2: Run loop, execute multiple calls
                VU2 ->> Promtail: Save []CallResult
                Promtail ->> Loki: Send batch<br/>when ready or timeout
            end
        end
        Scheduler ->> Scheduler: All segments done<br/>wait all responses<br/>test ends
        Profile(Test) ->> Grafana: Check alert groups<br/>FAIL or PASS the test
    end
```

Load workflow testing diagram:
```mermaid
---
title: Load testing workflow
---
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
Open [dashboard](http://localhost:3000/d/wasp/wasp-load-generator?orgId=1&refresh=5s&var-go_test_name=generator_healthcheck&var-gen_name=generator_healthcheck&var-branch=generator_healthcheck&var-commit=generator_healthcheck&from=now-5m&to=now)

`Gun` must implement this [interface](https://github.com/smartcontractkit/wasp/blob/master/wasp.go#L39)

## VUs test
- [test](https://github.com/smartcontractkit/wasp/blob/master/examples/simple_vu/main.go#L10)
- [vu](https://github.com/smartcontractkit/wasp/blob/master/examples/simple_vu/vu.go#L19)
```
cd examples/simple_vu
go run .
```
Open [dashboard](http://localhost:3000/d/wasp/wasp-load-generator?orgId=1&refresh=5s&var-go_test_name=generator_healthcheck&var-gen_name=generator_healthcheck&var-branch=generator_healthcheck&var-commit=generator_healthcheck&from=now-5m&to=now)

`VirtualUser` must implement this [interface](https://github.com/smartcontractkit/wasp/blob/master/wasp.go#L47)

## Usage in tests
- [test](https://github.com/smartcontractkit/wasp/blob/master/examples/go_test/main_test.go#L10)
- [gun](https://github.com/smartcontractkit/wasp/blob/master/examples/go_test/gun.go#L23)
```
cd examples/go_test
go test -v -count 1 .
```
Open [dashboard](http://localhost:3000/d/wasp/wasp-load-generator?orgId=1&refresh=5s&var-go_test_name=TestGenUsageWithTests&var-gen_name=generator_healthcheck&var-branch=generator_healthcheck&var-commit=generator_healthcheck&from=now-5m&to=now)

## Profile test (group multiple generators in parallel)
- [test](https://github.com/smartcontractkit/wasp/blob/master/examples/profiles/main_test.go#L11)
- [gun](https://github.com/smartcontractkit/wasp/blob/master/examples/profiles/gun.go#L23)
- [vu](https://github.com/smartcontractkit/wasp/blob/master/examples/profiles/vu.go#L19)
```
cd examples/profiles
go test -v -count 1 .
```
Open [dashboard](http://localhost:3000/d/wasp/wasp-load-generator?orgId=1&refresh=5s&var-go_test_name=TestProfile&var-gen_name=second%20API&var-gen_name=third%20API&var-gen_name=first%20API&var-branch=generator_healthcheck&var-commit=generator_healthcheck&from=now-5m&to=now)

## Scenario with simulating users behaviour
- [test](https://github.com/smartcontractkit/wasp/blob/master/examples/go_test/main_test.go#L15)
- [gun](https://github.com/smartcontractkit/wasp/blob/master/examples/go_test/gun.go#L23)
```
cd examples/scenario
go test -v -count 1 .
```
Open [dashboard](http://localhost:3000/d/wasp/wasp-load-generator?orgId=1&refresh=5s&var-go_test_name=TestScenario&var-gen_name=Two%20sequential%20calls%20scenario&var-branch=generator_healthcheck&var-commit=generator_healthcheck&from=now-5m&to=now)

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
- [definitions](https://github.com/smartcontractkit/wasp/blob/master/examples/alerts/main_test.go#L37)
- [wasp alerts](https://github.com/smartcontractkit/wasp/blob/master/examples/alerts/main_test.go#L73)
- [custom alerts](https://github.com/smartcontractkit/wasp/blob/master/examples/alerts/main_test.go#L82)
- [baseline NFR group test](https://github.com/smartcontractkit/wasp/blob/master/examples/alerts/main_test.go#L115)
- [stress NFR group test](https://github.com/smartcontractkit/wasp/blob/master/examples/alerts/main_test.go#L143)
```
cd examples/alerts
go test -v -count 1 -run TestBaselineRequirements
go test -v -count 1 -run TestStressRequirements
```
Open [alert groups](http://localhost:3000/alerting/groups)

Check [dashboard](http://localhost:3000/d/wasp/wasp-load-generator?orgId=1&refresh=5s&var-go_test_name=All&var-gen_name=All&var-branch=generator_healthcheck&var-commit=generator_healthcheck&from=now-5m&to=now), you can see per alert timeseries in the bottom
