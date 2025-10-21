# instructions

## build image

Build locally tagged images that are not pushed to any registry and are used exclusively within
the local Kubernetes cluster.

```shell
  make -C ../../../../ finder
  make -C ../../../../ reader-consumer
  make -C ../../../../ reader-worker
```

## kubernetes

Apply all recursively:

```bash
  kubectl create namespace youtube-chat-reader
  kubectl apply -R -f .
```
