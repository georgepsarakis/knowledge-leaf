package main

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/cors"
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

type requestLogger struct {
	middleware.LoggerInterface
	logger *zap.Logger
}

func (rl requestLogger) Print(v ...any) {
	rl.logger.Info("request completed", zap.Any("accessLog", v))
}

func newRedisClient(dsn string) *redis.Client {
	opts, err := redis.ParseURL(dsn)
	if err != nil {
		panic(err)
	}

	redisClient := redis.NewClient(opts)
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	if err := redisClient.Ping(ctx).Err(); err != nil {
		panic(err)
	}
	defer cancel()
	return redisClient
}

func main() {
	cfg := Configuration{}
	envconfig.MustProcess(context.Background(), &cfg)

	if cfg.UseRedis {
		_ = newRedisClient(cfg.RedisDSN)
	}

	logger, _ := zap.NewProduction()
	defer logger.Sync()
	r := chi.NewRouter()

	middleware.DefaultLogger = middleware.RequestLogger(
		&middleware.DefaultLogFormatter{
			Logger:  requestLogger{logger: logger},
			NoColor: true,
		},
	)

	r.Use(middleware.RequestID)
	r.Use(middleware.RealIP)
	r.Use(middleware.Logger)
	r.Use(cors.Handler(cors.Options{
		AllowedOrigins:   cfg.AllowedOrigins,
		AllowedMethods:   []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowedHeaders:   []string{"Accept", "Authorization", "Content-Type", "X-CSRF-Token"},
		AllowCredentials: false,
		MaxAge:           300, // Maximum value not ignored by any of major browsers
	}))
	r.Use(middleware.Recoverer)
	r.Use(func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Header().Add("Content-Type", "application/json")
			next.ServeHTTP(w, r)
		})
	})

	r.Use(middleware.Timeout(15 * time.Second))
	r.Use(middleware.AllowContentType("application/json"))
	r.Use(middleware.StripSlashes)
	r.Use(middleware.CleanPath)

	// Routes
	r.Get("/trivia/random", func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()

		loggerFields := []zap.Field{
			zap.String("requestID", middleware.GetReqID(ctx)),
			zap.String("httpMethod", http.MethodGet),
			zap.String("operation", "trivia/random"),
		}
		logger = logger.With(loggerFields...)

		// Search for a Wikipedia article
		summaries, err := randomizeArticle(ctx)
		if err != nil {
			logger.Error(err.Error(), zap.Error(err))
			http.Error(w, "request failed", http.StatusInternalServerError)
			return
		}
		b, err := json.Marshal(RandomTriviaResponse{Results: summaries})
		if err != nil {
			logger.Error(err.Error(),
				zap.Error(err),
				zap.String("operationDetail", "jsonMarshal"))
			http.Error(w, "request failed", http.StatusInternalServerError)
			return
		}
		if _, err := w.Write(b); err != nil {
			logger.Error(err.Error(),
				zap.Error(err),
				zap.String("operationDetail", "responseWrite"))
			http.Error(w, "request failed", http.StatusInternalServerError)
			return
		}
	})

	if err := http.ListenAndServe(fmt.Sprintf(":%d", cfg.Port), r); err != nil {
		logger.Error(err.Error())
		os.Exit(1)
	}
}
