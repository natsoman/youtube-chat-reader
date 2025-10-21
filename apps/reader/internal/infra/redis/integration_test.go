//go:build integration

package redis_test

import (
	"context"
	"log"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/docker/go-connections/nat"
	goredis "github.com/redis/go-redis/v9"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/wait"

	infraredis "github.com/natsoman/youtube-chat-reader/apps/reader/internal/infra/redis"
)

var (
	_redisC testcontainers.Container
	_client *goredis.ClusterClient
	_locker *infraredis.Locker
)

func TestMain(m *testing.M) {
	ctx := context.Background()

	redisC, err := testcontainers.Run(ctx, "redis:8.2-alpine",
		testcontainers.WithCmdArgs(
			"redis-server",
			"--cluster-enabled", "yes",
			"--appendonly", "no",
			"--save", "",
		),
		testcontainers.WithWaitStrategy(
			wait.ForLog("Ready to accept connections").
				WithStartupTimeout(15*time.Second),
		),
	)

	defer func() {
		_ = testcontainers.TerminateContainer(redisC)
	}()

	// Get Redis connection details
	host, err := redisC.Host(ctx)
	if err != nil {
		log.Fatalf("Failed to get container host: %v", err)
	}

	port, err := redisC.MappedPort(ctx, nat.Port("6379/tcp"))
	if err != nil {
		log.Fatalf("Failed to get mapped port: %v", err)
	}

	// Create a standard client to initialize the cluster
	addr := host + ":" + port.Port()
	tempClient := goredis.NewClient(&goredis.Options{Addr: addr})

	// Configure the cluster to announce the host address
	if err = tempClient.ConfigSet(ctx, "cluster-announce-ip", host).Err(); err != nil {
		log.Fatalf("Failed to set cluster-announce-ip: %v", err)
	}

	if err = tempClient.ConfigSet(ctx, "cluster-announce-port", port.Port()).Err(); err != nil {
		log.Fatalf("Failed to set cluster-announce-port: %v", err)
	}

	// Assign all slots to this node (0-16383)
	slots := make([]int, 16384)
	for i := 0; i < 16384; i++ {
		slots[i] = i
	}

	if err = tempClient.ClusterAddSlots(ctx, slots...).Err(); err != nil {
		log.Fatalf("Failed to add slots: %v", err)
	}

	// Wait for cluster to be ready
	for i := 0; i < 20; i++ {
		info, err := tempClient.ClusterInfo(ctx).Result()
		if err == nil && strings.Contains(info, "cluster_state:ok") {
			break
		}
		if i == 19 {
			log.Fatalf("Cluster did not become ready in time. Info: %v, Error: %v", info, err)
		}
		time.Sleep(time.Millisecond * 500)
	}

	tempClient.Close()

	_redisC = redisC

	_client = goredis.NewClusterClient(&goredis.ClusterOptions{Addrs: []string{addr}})
	defer func() { _ = _client.Close() }()

	_locker, err = infraredis.NewLocker(_client)
	if err != nil {
		log.Fatalf("Failed to create locker: %v", err)
	}

	os.Exit(m.Run())
}
