# instructions

```shell
    kubectl create namespace mongo
    kubectl apply -f service.yaml
    kubectl apply -f statefulset.yaml
```

```shell
    chmod +x ./init-replicaset.sh && ./init-replicaset.sh
```

## prometheus-mongodb-exporter

```shell
    for i in 0 1 2; do
      helm upgrade -i prometheus-mongodb-exporter-$i \
        oci://ghcr.io/prometheus-community/charts/prometheus-mongodb-exporter \
        -f prometheus-mongodb-exporter/values-$i.yaml -n mongo
    done
```
