# instructions

## telepresence

Use [Telepresence](https://telepresence.io) to enable access to Kubernetes services from your Docker host.

```bash
    telepresence helm install
    kubectl rollout status deploy/traffic-manager -n ambassador
    telepresence connect
```

## components

Follow the instructions provided in the respective component directories:

- Kafka: [README.md](./kafka/README.md)
- Mongo: [README.md](./mongodb/README.md)
- Redis: [README.md](./redis/README.md)
- OTEL:  [README.md](./otel/README.md)
- Services: [README.md](./youtube-chat-reader/README.md)
