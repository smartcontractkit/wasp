.PHONY: test
test:
	go test -v -count 1 `go list ./... | grep -v examples` -run TestSmoke

.PHONY: test_race
test_race:
	go test -v -race -count 1 `go list ./... | grep -v examples` -run TestSmoke

.PHONY: test+cover
test_cover:
	go test -v -coverprofile cover.out -count 1 `go list ./... | grep -v examples` -run TestSmoke
	go tool cover -html cover.out

.PHONY: test
test_loki:
	go test -v -count 1 `go list ./... | grep -v examples` -run TestPerfRenderLoki

.PHONY: test
test_pyro:
	go test -v -run TestPyroscopeLocalTrace -trace trace.out

.PHONY: dashboard
dashboard:
	go run dashboard/dashboard.go

.PHONY: start
start:
	docker compose -f compose/docker-compose.yaml up -d
	sleep 5 && curl -X POST -H "Content-Type: application/json" -d '{"name":"test", "role": "Admin"}' http://localhost:3000/api/auth/keys | jq .key

.PHONY: stop
stop:
	docker compose -f compose/docker-compose.yaml down -v

.PHONY: pyro_start
pyro_start:
	docker compose -f compose/pyroscope-compose.yaml up -d

.PHONY: pyro_stop
pyro_stop:
	docker compose -f compose/pyroscope-compose.yaml down -v

.PHONY: lint
lint:
	golangci-lint --color=always run -v
