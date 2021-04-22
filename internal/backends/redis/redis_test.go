package redis_test

import (
	"context"
	"testing"

	"github.com/OpenSlides/openslides-vote-service/internal/backends/redis"
	"github.com/OpenSlides/openslides-vote-service/internal/backends/test"
	"github.com/ory/dockertest/v3"
)

func startRedis(t *testing.T) (string, func()) {
	pool, err := dockertest.NewPool("")
	if err != nil {
		t.Fatalf("Could not connect to docker: %s", err)
	}

	resource, err := pool.Run("redis", "6.2", nil)
	if err != nil {
		t.Fatalf("Could not start resource: %s", err)
	}

	return resource.GetPort("6379/tcp"), func() {
		if err = pool.Purge(resource); err != nil {
			t.Fatalf("Could not purge resource: %s", err)
		}
	}
}

func TestVote(t *testing.T) {
	port, close := startRedis(t)
	defer close()

	r := redis.New("localhost:" + port)
	r.Wait(context.Background(), nil)
	t.Logf("Redis port: %s", port)

	test.Backend(t, r)
}
