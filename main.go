package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/cors"
	"go.uber.org/zap"

	"knowledgeleaf/app"
	"knowledgeleaf/externalapi/wikipedia"
)

type requestLogger struct {
	middleware.LoggerInterface
	logger *zap.Logger
}

func (rl requestLogger) Print(v ...any) {
	rl.logger.Info("request completed", zap.Any("accessLog", v))
}

func main() {
	application, cleanup, err := app.New()
	if err != nil {
		panic(err)
	}
	defer cleanup()

	r := chi.NewRouter()

	middleware.DefaultLogger = middleware.RequestLogger(
		&middleware.DefaultLogFormatter{
			Logger:  requestLogger{logger: application.Logger},
			NoColor: true,
		},
	)

	r.Use(middleware.RequestID)
	r.Use(middleware.RealIP)
	r.Use(middleware.Logger)
	r.Use(cors.Handler(cors.Options{
		AllowedOrigins:   application.Cfg.AllowedOrigins,
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
	r.Use(func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			ctx := app.WithLogger(r.Context(), application.Logger)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	})

	triviaBackend := NewRandomTriviaBackend(application)

	// Routes
	r.Get("/trivia/random", func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		logger := app.LoggerFromContext(ctx)
		loggerFields := []zap.Field{
			zap.String("requestID", middleware.GetReqID(ctx)),
			zap.String("httpMethod", http.MethodGet),
			zap.String("operation", "trivia/random"),
		}
		logger = logger.With(loggerFields...)

		// Search for a Wikipedia article
		summaries, err := randomizeArticle(ctx, triviaBackend)
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
	r.Get("/trivia/stats", func(w http.ResponseWriter, r *http.Request) {
		// TODO: return total database size count & views
	})
	r.Get("/on-this-day/events", func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		logger := app.LoggerFromContext(ctx)
		loggerFields := []zap.Field{
			zap.String("requestID", middleware.GetReqID(ctx)),
			zap.String("httpMethod", http.MethodGet),
			zap.String("operation", "on-this-day/events"),
		}
		logger = logger.With(loggerFields...)

		client := wikipedia.NewClient()
		events, err := client.OnThisDay(ctx)
		if err != nil {
			logger.Error(err.Error(), zap.Error(err))
			http.Error(w, "request failed", http.StatusInternalServerError)
			return
		}
		var resp EventsOnThisDayResponse
		for _, ev := range events.Events {
			mainPage := ev.Pages[0]
			var references []OnThisDayEventReference
			if len(ev.Pages) > 1 {
				for _, p := range ev.Pages[1:] {
					references = append(references, OnThisDayEventReference{
						Title: p.Title,
						URL:   p.ContentUrls.Desktop.Page,
					})
				}
			}
			resp.Titles = append(resp.Titles, OnThisDayEvent{
				Title:      ev.Text,
				ShortTitle: mainPage.Titles.Normalized,
				Image: Image{
					URL:    mainPage.Thumbnail.Source,
					Width:  mainPage.Thumbnail.Width,
					Height: mainPage.Thumbnail.Height,
				},
				Description: mainPage.Description,
				Extract:     mainPage.Extract,
				URL:         mainPage.ContentUrls.Desktop.Page,
				References:  references,
			})
		}
		b, err := json.Marshal(resp)
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

	if err := http.ListenAndServe(fmt.Sprintf(":%d", application.Cfg.Port), r); err != nil {
		application.Logger.Error(err.Error())
		os.Exit(1)
	}
}
