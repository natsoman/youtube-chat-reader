# instructions

## Deploy Strimzi using installation files

Follow https://strimzi.io/quickstarts/#deploy-strimzi-using-installation-files

```shell
  kubectl create namespace kafka
  kubectl create -f 'https://strimzi.io/install/latest?namespace=kafka' -n kafka
  kubectl get pod -n kafka --watch
```

## Create an Apache Kafka cluster

```shell
  kubectl apply -f kafka-node-pool.yaml -f kafka.yaml -n kafka
```

Connect to Kafka cluster using:

```text
youtube-dual-role-0.youtube-kafka-brokers.kafka.svc.cluster.local:9092
youtube-dual-role-1.youtube-kafka-brokers.kafka.svc.cluster.local:9092
youtube-dual-role-2.youtube-kafka-brokers.kafka.svc.cluster.local:9092
```
