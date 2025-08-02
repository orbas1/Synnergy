# Monitoring

This directory contains assets for observing Synnergy components.

## Prometheus Metrics

Use the `core.HealthLogger` to expose runtime metrics. After creating the logger,
call `RunMetricsCollector` to capture metrics on an interval and start the metrics endpoint (example uses the standard `context` and `time` packages):

```go
logger, _ := core.NewHealthLogger(ledger, node, coin, txpool, "health.log")
ctx, cancel := context.WithCancel(context.Background())
go logger.RunMetricsCollector(ctx, time.Minute)
srv, _ := logger.StartMetricsServer(":9090")
defer logger.ShutdownMetricsServer(ctx, srv)
defer cancel()
```

## Grafana Dashboards

- `grafana/dashboard-node.json` – core node metrics such as block height and peers.
- `grafana/dashboard-gui.json` – GUI usage metrics like page load times.
- `grafana/dashboard-cli.json` – CLI command counts and latency.

Import these JSON files into Grafana to visualize system behaviour.
