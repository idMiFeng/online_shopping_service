package redis

import (
	"context"
	"fmt"
	"github.com/idMiFeng/stock_service/config"

	"github.com/go-redis/redis/v8"
	"github.com/go-redsync/redsync/v4"
	"github.com/go-redsync/redsync/v4/redis/goredis/v8"
)

// redsync -> https://github.com/go-redsync/redsync

var (
	rc *redis.Client
	Rs *redsync.Redsync
)

func Init(cfg *config.RedisConfig) error {
	rc = redis.NewClient(&redis.Options{
		Addr:     fmt.Sprintf("%s:%d", cfg.Host, cfg.Port),
		Password: cfg.Password, // 密码
		DB:       cfg.DB,       // 数据库
		PoolSize: cfg.PoolSize, // 连接池大小
	})
	err := rc.Ping(context.Background()).Err()
	if err != nil {
		return err
	}
	pool := goredis.NewPool(rc) // or, pool := redigo.NewPool(...)

	// Create an instance of redisync to be used to obtain a mutual exclusion
	// lock.
	Rs = redsync.New(pool)
	return nil
}
