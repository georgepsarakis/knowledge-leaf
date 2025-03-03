package app

import (
	"context"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
	"github.com/sethvargo/go-envconfig"
	"go.uber.org/zap"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

type Configuration struct {
	Port             int           `env:"PORT,default=4000"`
	AllowedOrigins   []string      `env:"ALLOWED_ORIGINS,default=http://localhost"`
	RedisDSN         string        `env:"REDIS_DSN,default=redis://localhost:6379"`
	UseRedis         bool          `env:"USE_REDIS,default=false"`
	RequestTimeout   time.Duration `env:"REQUEST_TIMEOUT,default=30s"`
	PostgresHost     string        `env:"POSTGRES_HOST,default=localhost"`
	PostgresPort     int           `env:"POSTGRES_PORT,default=5432"`
	PostgresUser     string        `env:"POSTGRES_USER,default=knowledge_leaf"`
	PostgresPassword string        `env:"POSTGRES_PASSWORD"`
	PostgresDatabase string        `env:"POSTGRES_DATABASE,default=knowledge_leaf"`
	PostgresEnabled  bool          `env:"POSTGRES_ENABLED,default=false"`
}

type App struct {
	Cfg                Configuration
	RedisClient        *redis.Client
	Logger             *zap.Logger
	PostgresConnection *gorm.DB
}

func New() (App, func() error, error) {
	app := App{}
	cfg := Configuration{}
	envconfig.MustProcess(context.Background(), &cfg)
	app.Cfg = cfg

	appLogger, _ := zap.NewProduction()
	app.Logger = appLogger

	if cfg.UseRedis {
		app.Logger.Info("using Redis database")
		rc, err := newRedisClient(cfg.RedisDSN)
		if err != nil {
			return app, nil, err
		}
		app.RedisClient = rc
	}
	if app.Cfg.PostgresEnabled {
		dsn := fmt.Sprintf("user=%s password=%s host=%s port=%d dbname=%s",
			app.Cfg.PostgresUser,
			app.Cfg.PostgresPassword,
			app.Cfg.PostgresHost,
			app.Cfg.PostgresPort,
			app.Cfg.PostgresDatabase,
		)
		db, err := gorm.Open(
			postgres.New(postgres.Config{DSN: dsn}), &gorm.Config{
				Logger: logger.Default.LogMode(logger.Silent),
			})
		if err != nil {
			return app, nil, err
		}
		app.PostgresConnection = db
	}

	return app, func() error {
		return app.Logger.Sync()
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
