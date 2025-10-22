# instructions

## prerequisites

```shell
  helm repo add open-telemetry https://open-telemetry.github.io/opentelemetry-helm-charts
  helm repo add grafana https://grafana.github.io/helm-charts
  helm repo add prometheus-community https://prometheus-community.github.io/helm-charts
  helm repo update
```

```shell
  kubectl create namespace observability
```

## Open Telemetry collector

The OpenTelemetry Collector will receive telemetry data and forward it to the appropriate backends:
- Logs:  Loki
- Metrics: Prometheus
- Traces: Tempo

```shell
  helm upgrade -i otel-collector open-telemetry/opentelemetry-collector -f otel-collector-values.yaml -n observability
```

## loki

```shell
  helm upgrade -i loki grafana/loki -f loki-values.yaml -n observability
```

## tempo

```shell
  helm upgrade -i tempo grafana/tempo -n observability
```

[See more](https://github.com/grafana/helm-charts/tree/main/charts/tempo)

## prometheus

```shell
  helm upgrade -i prometheus prometheus-community/prometheus -f prometheus-values.yaml -n observability
```

Access Prometheus UI at: http://localhost:9090

## grafana

```shell
  helm upgrade -i grafana grafana/grafana -f grafana-values.yaml -n observability
```

Access Grafana at: http://grafana.observability.svc.cluster.local/

- Username: `admin`
- Password:
```shell
  kubectl get secret -n observability grafana -o jsonpath="{.data.admin-password}" | base64 --decode ; echo
```
