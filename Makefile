.PHONY: test
test:
	go test -v ./... -run TestSmoke

.PHONY: test
test_loki:
	go test -v -count 1 ./... -run TestRender

.PHONY: test
test_pyro:
	go test -v -run TestPyroscopeLocalTrace -trace trace.out

.PHONY: dashboard
dashboard:
	go run dashboard/dashboard.go

.PHONY: start
start:
	docker compose -f compose/docker-compose.yaml up -d

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
	./bin/golangci-lint --color=always run -v

.PHONY: install_deps
install_deps:
	curl -sSfL https://raw.githubusercontent.com/golangci/golangci-lint/master/install.sh | sh -s v1.51.2
