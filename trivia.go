package main

import (
	"context"
	"embed"
	"encoding/json"
	"errors"
	"fmt"
	"math/rand"

	"go.uber.org/zap"
	"golang.org/x/sync/errgroup"
	"gorm.io/gorm/clause"

	"knowledgeleaf/app"
	"knowledgeleaf/database"
	"knowledgeleaf/externalapi/wikipedia"
)

// https://dumps.wikimedia.org/enwiki/latest/
//
//go:embed knowledgebase/samples/wikipedia-article-list-samples-*.json
var knowledgeBaseFS embed.FS

var wikipediaArticleTitles = func() []string {
	var titles []string
	for _, pathSuffix := range []string{"1"} {
		var elements []string
		b, err := knowledgeBaseFS.ReadFile(
			fmt.Sprintf("knowledgebase/samples/wikipedia-article-list-samples-%s.json", pathSuffix))
		if err != nil {
			panic(err)
		}
		if err := json.Unmarshal(b, &elements); err != nil {
			panic(err)
		}
		titles = append(titles, elements...)
	}

	return titles
}()

type RandomTriviaBackend struct {
	application app.App
	titleCount  int
}

func NewRandomTriviaBackend(application app.App) *RandomTriviaBackend {
	return &RandomTriviaBackend{application: application}
}

func (b *RandomTriviaBackend) RandomTitle(ctx context.Context) (string, error) {
	if b.application.Cfg.PostgresEnabled {
		return fetchRandomArticleTitle(ctx, b)
	}

	if !b.application.Cfg.UseRedis {
		if b.titleCount == 0 {
			b.titleCount = len(wikipediaArticleTitles)
		}
		return wikipediaArticleTitles[rand.Intn(b.titleCount)], nil
	}

	if b.titleCount == 0 {
		// TODO: reuse key names between Loader and Fetcher
		cmd := b.application.RedisClient.SCard(ctx, "datasource:wikipedia")
		if cmd.Err() != nil {
			return "", cmd.Err()
		}
		b.titleCount = int(cmd.Val())
		b.application.Logger.Info(fmt.Sprintf("found %d titles in Redis DB", b.titleCount))
	}
	title, err := b.application.RedisClient.SRandMember(ctx, "datasource:wikipedia").Result()
	if err != nil {
		return "", err
	}
	return title, nil
}

const maxTries = 2

func randomizeArticle(ctx context.Context, triviaBackend *RandomTriviaBackend) ([]WikiSummary, error) {
	var (
		summaryResp wikipedia.RestV1SummaryResponse
		categories  []string
	)
	for iter := 0; iter < maxTries; iter++ {
		subj, err := triviaBackend.RandomTitle(ctx)
		if err != nil {
			return nil, err
		}
		client := wikipedia.NewClient()

		group, ctx := errgroup.WithContext(ctx)
		group.Go(func() error {
			summary, err := client.GetSummary(ctx, subj)
			if err != nil {
				return err
			}
			summaryResp = summary
			return nil
		})
		group.Go(func() error {
			var err error
			categories, err = client.Categories(ctx, subj)
			return err
		})
		if err := group.Wait(); err != nil {
			if errors.Is(err, wikipedia.ErrNotFound) && iter < maxTries-1 {
				logger := app.LoggerFromContext(ctx)
				logger.Info("page not found - retrying", zap.String("title", subj))
				continue
			}
			return nil, err
		}
		break
	}

	var summaries []WikiSummary
	summaries = append(summaries,
		WikiSummary{
			Title:   summaryResp.Title,
			Summary: summaryResp.Extract,
			Metadata: WikiSummaryMetadata{
				Description: summaryResp.Description,
				URL:         summaryResp.ContentUrls.Desktop.Page,
				Image: Image{
					URL:    summaryResp.Thumbnail.Source,
					Width:  summaryResp.Thumbnail.Width,
					Height: summaryResp.Thumbnail.Height,
				},
			},
			Categories: categories,
		},
	)
	return summaries, nil
}

const numericIDSequence = "wikipedia_titles_numeric_id_seq"

func fetchRandomArticleTitle(ctx context.Context, triviaBackend *RandomTriviaBackend) (string, error) {
	tx := triviaBackend.application.PostgresConnection.WithContext(ctx)
	var n int64
	err := tx.Raw(fmt.Sprintf("SELECT LAST_VALUE FROM %s", numericIDSequence)).Scan(&n).Error
	if err != nil {
		return "", err
	}
	bucket := rand.Intn(int(n)) + 1
	title := database.WikipediaTitle{}
	err = tx.Clauses(clause.OrderBy{
		Expression: clause.Expr{SQL: "RANDOM()"},
	}).First(&title, "numeric_id = ?", bucket).Error
	if err != nil {
		return "", err
	}
	return title.Title, nil
}
