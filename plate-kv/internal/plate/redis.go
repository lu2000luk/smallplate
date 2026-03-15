package plate

import (
	"context"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
)

type Dependencies struct {
	Config          Config
	Redis           *redis.Client
	PubSub          *redis.Client
	AuthCache       *AuthCache
	Manager         *ManagerClient
	CommandRegistry map[string]CommandSpec
}

func NewDependencies(cfg Config) (*Dependencies, error) {
	cache, err := NewAuthCache(cfg.AuthCacheSize)
	if err != nil {
		return nil, err
	}

	redisClient, pubSubClient, err := NewRedisClients(cfg)
	if err != nil {
		return nil, err
	}

	deps := &Dependencies{
		Config:          cfg,
		Redis:           redisClient,
		PubSub:          pubSubClient,
		AuthCache:       cache,
		CommandRegistry: NewCommandRegistry(),
	}
	deps.Manager = NewManagerClient(cfg, cache, redisClient)
	return deps, nil
}

func (d *Dependencies) Close() error {
	var errs []error
	if d.PubSub != nil {
		if err := d.PubSub.Close(); err != nil {
			errs = append(errs, err)
		}
	}
	if d.Redis != nil {
		if err := d.Redis.Close(); err != nil {
			errs = append(errs, err)
		}
	}
	if len(errs) == 0 {
		return nil
	}
	return fmt.Errorf("close errors: %v", errs)
}

func NewRedisClients(cfg Config) (*redis.Client, *redis.Client, error) {
	options, err := redis.ParseURL(cfg.DBURL)
	if err != nil {
		return nil, nil, err
	}
	commandOptions := *options
	pubSubOptions := *options
	pubSubOptions.PoolSize = max(options.PoolSize, 16)

	commandClient := redis.NewClient(&commandOptions)
	pubSubClient := redis.NewClient(&pubSubOptions)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := commandClient.Ping(ctx).Err(); err != nil {
		_ = commandClient.Close()
		_ = pubSubClient.Close()
		return nil, nil, err
	}
	if err := pubSubClient.Ping(ctx).Err(); err != nil {
		_ = commandClient.Close()
		_ = pubSubClient.Close()
		return nil, nil, err
	}
	return commandClient, pubSubClient, nil
}

func max(a int, b int) int {
	if a > b {
		return a
	}
	return b
}
