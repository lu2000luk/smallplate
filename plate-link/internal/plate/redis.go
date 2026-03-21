package plate

import (
	"context"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
)

type Dependencies struct {
	Config    Config
	Redis     *redis.Client
	AuthCache *AuthCache
	Manager   *ManagerClient
	Links     *LinkStore
}

func NewDependencies(cfg Config) (*Dependencies, error) {
	cache, err := NewAuthCache(cfg.AuthCacheSize)
	if err != nil {
		return nil, err
	}

	redisClient, err := NewRedisClient(cfg)
	if err != nil {
		return nil, err
	}

	deps := &Dependencies{
		Config:    cfg,
		Redis:     redisClient,
		AuthCache: cache,
	}
	deps.Links = NewLinkStore(cfg, redisClient)
	deps.Manager = NewManagerClient(cfg, cache, redisClient, deps.Links)
	return deps, nil
}

func (d *Dependencies) Close() error {
	if d.Redis != nil {
		return d.Redis.Close()
	}
	return nil
}

func NewRedisClient(cfg Config) (*redis.Client, error) {
	options, err := redis.ParseURL(cfg.DBURL)
	if err != nil {
		return nil, err
	}

	client := redis.NewClient(options)
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := client.Ping(ctx).Err(); err != nil {
		_ = client.Close()
		return nil, err
	}
	return client, nil
}

func CloseErrors(errs []error) error {
	if len(errs) == 0 {
		return nil
	}
	return fmt.Errorf("close errors: %v", errs)
}
