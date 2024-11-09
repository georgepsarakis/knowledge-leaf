package app

import (
	"context"
	"time"

	"github.com/redis/go-redis/v9"
	"github.com/sethvargo/go-envconfig"
	"go.uber.org/zap"
)

type Configuration struct {
	Port           int      `env:"PORT,default=4000"`
	AllowedOrigins []string `env:"ALLOWED_ORIGINS,default=http://localhost"`
	RedisDSN       string   `env:"REDIS_DSN,default=redis://localhost:6379"`
	UseRedis       bool     `env:"USE_REDIS,default=false"`
}

type App struct {
	Cfg         Configuration
	RedisClient *redis.Client
	Logger      *zap.Logger
}

func New() (App, func() error, error) {
	app := App{}
	cfg := Configuration{}
	envconfig.MustProcess(context.Background(), &cfg)
	app.Cfg = cfg

	logger, _ := zap.NewProduction()
	app.Logger = logger

	if cfg.UseRedis {
		logger.Info("using Redis database")
		rc, err := newRedisClient(cfg.RedisDSN)
		if err != nil {
			return app, nil, err
		}
		app.RedisClient = rc
	}

	return app, func() error {
		return logger.Sync()
	}, nil
}

func newRedisClient(dsn string) (*redis.Client, error) {
	opts, err := redis.ParseURL(dsn)
	if err != nil {
		return nil, err
	}

	redisClient := redis.NewClient(opts)
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	if err := redisClient.Ping(ctx).Err(); err != nil {
		return nil, err
	}
	return redisClient, nil
}
