# instructions

Run the following comment to install Redis cluster to your local Kubernetes cluster:

```shell
  helm upgrade --install redis-cluster \
    --set usePassword=false \
    --namespace redis \
    --create-namespace \
    oci://registry-1.docker.io/bitnamicharts/redis-cluster
```

Access Redis instances using:

```text
redis-cluster-0.redis-cluster-headless.redis.svc.cluster.local:6379
redis-cluster-1.redis-cluster-headless.redis.svc.cluster.local:6379
redis-cluster-2.redis-cluster-headless.redis.svc.cluster.local:6379
redis-cluster-3.redis-cluster-headless.redis.svc.cluster.local:6379
redis-cluster-4.redis-cluster-headless.redis.svc.cluster.local:6379
redis-cluster-5.redis-cluster-headless.redis.svc.cluster.local:6379
```

More: [Bitnami package for Redis®](https://github.com/bitnami/charts/tree/main/bitnami/redis)
