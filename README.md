# YouTube Live Chat Reader

[![CI](https://github.com/natsoman/youtube-chat-reader/actions/workflows/ci.yaml/badge.svg)](https://github.com/natsoman/youtube-chat-reader/actions/workflows/ci.yaml)
[![codecov](https://codecov.io/gh/natsoman/youtube-chat-reader/graph/badge.svg?token=QXORZL6UE8)](https://codecov.io/gh/natsoman/youtube-chat-reader)
[![Go Report Card](https://goreportcard.com/badge/github.com/natsoman/youtube-chat-reader/apps/reader)](https://goreportcard.com/report/github.com/natsoman/youtube-chat-reader/apps/reader)

A high-performance system for efficiently reading YouTube Live Chat messages.

## 🚀 Features

- **Real-time Processing**: Reading YouTube Live Chat messages with the most efficient method, [streamList](https://developers.google.com/youtube/v3/live/docs/liveChatMessages/streamList)
- **Scalable & High Available**: Horizontally scalable reading workers; if one fails, another one takes over almost immediately
- **Observable**: Comprehensive metrics, tracing, and logging
- **Kubernetes-Native**: Designed to run in containerized environments

## 🏗️ Architecture

The system consists of two main components:

### 1. Finder Service
- Periodically checks configured YouTube channels for upcoming live streams
- Emits events when new live streams are discovered
- Runs as a Kubernetes CronJob for scheduled execution

### 2. Reader Service
A multi-binary service with the following components:

#### Consumer
- Listens for newly discovered live stream events
- Creates the initial live stream progress state
- Runs as a Kubernetes [Deployment](./deploy/local/k8s/youtube-chat-reader/reader/consumer/deployment.yaml)

#### Worker
- Reads and stores live chat messages using YouTube’s [streamList](https://developers.google.com/youtube/v3/live/docs/liveChatMessages/streamList)
- Distributed locking with Etcd to ensure exactly-once live stream processing
- Runs as a Kubernetes [Deployment](./deploy/local/k8s/youtube-chat-reader/reader/worker/deployment.yaml)

## 🚀 Quick Start

### Prerequisites

- Kubernetes cluster (e.g., Docker Desktop)
- Helm 3+
- kubectl

### Installation

1. Clone the repository:
   ```bash
   git clone https://github.com/natsoman/youtube-chat-reader.git
   cd youtube-chat-reader
   ```

2. Deploy to your Kubernetes cluster:
   ```bash
   # Navigate to the deployment directory and follow instructions
   cd deploy/local/k8s/
   ```

## ⚙️ Configuration

Configuration is managed through Kubernetes resources:

### Common Configuration
- [ConfigMap](./deploy/local/k8s/youtube-chat-reader/configmap.yaml)
- [Secrets](./deploy/local/k8s/youtube-chat-reader/secret.yaml)

### Component-Specific Configuration

#### Finder
- [CronJob](./deploy/local/k8s/youtube-chat-reader/finder/cronjob.yaml)
- [Secrets](./deploy/local/k8s/youtube-chat-reader/finder/secret.yaml)

#### Reader Worker
- [Deployment](./deploy/local/k8s/youtube-chat-reader/reader/worker/deployment.yaml)
- [Secrets](./deploy/local/k8s/youtube-chat-reader/reader/worker/secret.yaml)

#### Reader Consumer
- [Deployment](./deploy/local/k8s/youtube-chat-reader/reader/consumer/deployment.yaml)

## Coverage

![Coverage](https://codecov.io/gh/natsoman/youtube-chat-reader/graphs/icicle.svg?token=QXORZL6UE8)
