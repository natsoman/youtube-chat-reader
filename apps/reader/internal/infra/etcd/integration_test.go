//go:build integration

package etcd_test

import (
	"context"
	"log"
	"os"
	"testing"
	"time"

	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/modules/etcd"
	clientv3 "go.etcd.io/etcd/client/v3"
)

var (
	_client *clientv3.Client
)

func TestMain(m *testing.M) {
	ctx := context.Background()

	etcdContainer, err := etcd.Run(ctx, "gcr.io/etcd-development/etcd:v3.5.14",
		etcd.WithNodes("etcd-1", "etcd-2", "etcd-3"),
	)
	defer func() {
		if err = testcontainers.TerminateContainer(etcdContainer); err != nil {
			log.Fatal(err)
		}
	}()
	if err != nil {
		log.Fatal(err)
	}

	clientEndpoint, err := etcdContainer.ClientEndpoint(ctx)
	if err != nil {
		log.Fatal(err)
	}

	client, err := clientv3.New(clientv3.Config{
		Endpoints:   []string{clientEndpoint},
		DialTimeout: 5 * time.Second,
	})
	if err != nil {
		log.Fatal(err)
	}
	defer func() {
		_ = _client.Close()
	}()

	_client = client

	os.Exit(m.Run())
}
