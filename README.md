# xk6-gtp

A client extension for interacting with the GTP protocol in your k6 tests.


## Preparation

Required packages and tools:

- [mise](https://github.com/jdx/mise)

Install the tools required for development:

```shell=
make install-dev-pkg
```

## Build

```shell=
make install-go-tools
make build
```

## Running Tests

```shell
./out/bin/xk6 run example/echo-stress.js

./out/bin/pgw
```

## Grafana Dashboard

A ready-to-import dashboard covering the extension's six k6 metrics
(`gtpv2_req_total`, `gtpv2_req_duration`, `gtpv2_resp_cause`,
`gtpv2_timeout_total`, `gtpv2_send_error_total`,
`gtpv2_conn_reconnect_total`) lives at
[`dashboards/xk6-gtp-overview.json`](dashboards/xk6-gtp-overview.json).

The dashboard targets metric names produced by
[`xk6-output-prometheus-remote`](https://github.com/grafana/xk6-output-prometheus-remote)
(each k6 metric surfaces as `k6_<name>` in Prometheus). Run k6 with the
Prometheus remote-write output and import the JSON into Grafana:

```sh
# 1. Run k6 with prometheus remote-write (adjust endpoint / auth as needed).
K6_PROMETHEUS_RW_SERVER_URL=http://prometheus:9090/api/v1/write \
  ./out/bin/xk6 run -o experimental-prometheus-rw examples/troubleshoot-session-health.js

# 2. Import dashboards/xk6-gtp-overview.json from the Grafana UI
#    (Dashboards → New → Import → Upload JSON file).
```

## Supported Scenarios

### GTPv2-C

- [x] Node monitoring (Echo Request/Echo Response)
- [x] Create Session  (Create Session Request/Create Session Response)
  - [x] sgw->pgw scenario
- [x] Delete Session (Delete Session Request/Delete Session Response)
  - [x] sgw->pgw scenario
- [x] Modify Bearer (Modify Bearer Request/Modify Bearer Response)
  - [x] sgw->pgw scenario
- [ ] Delete Bearer (Delete Bearer Request/Delete Bearer Response)

## Special Thanks

This PoC takes full advantage of [go-gtp](https://github.com/wmnsk/go-gtp). Thanks to @wmnsk and all the developers.

## Developer Settings

```shell
# Format, lint, commit message validation, etc.
pre-commit install

# Mob programming
co-author hook > .git/hooks/prepare-commit-msg
chmod +x .git/hooks/prepare-commit-msg

# Create Docker image
make docker-build
make docker-release
```
