package main

import (
	"context"
	"embed"
	"encoding/json"
	"fmt"
	"math/rand"

	"knowledgeleaf/app"
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
	if !b.application.Cfg.UseRedis {
		if b.titleCount == 0 {
			b.titleCount = len(wikipediaArticleTitles)
		}
		return wikipediaArticleTitles[rand.Intn(b.titleCount)], nil
	}

	if b.titleCount == 0 {
		// TODO: reuse key names between Loader and Fetcher
		cmd := b.application.RedisClient.ZCard(ctx, "datasource:wikipedia")
		if cmd.Err() != nil {
			return "", cmd.Err()
		}
		b.titleCount = int(cmd.Val())
		b.application.Logger.Info(fmt.Sprintf("found %d titles in Redis DB", b.titleCount))
	}
	index := int64(rand.Intn(b.titleCount))
	titles, err := b.application.RedisClient.ZRange(ctx, "datasource:wikipedia", index, index).Result()
	if err != nil {
		return "", err
	}
	return titles[0], nil
}

func randomizeArticle(ctx context.Context, triviaBackend *RandomTriviaBackend) ([]WikiSummary, error) {
	subj, err := triviaBackend.RandomTitle(ctx)
	if err != nil {
		return nil, err
	}
	client := wikipedia.NewClient()
	summary, err := client.GetSummary(ctx, subj)
	if err != nil {
		return nil, err
	}

	var summaries []WikiSummary
	summaries = append(summaries,
		WikiSummary{
			Title:   summary.Title,
			Summary: summary.Extract,
			Metadata: WikiSummaryMetadata{
				Description: summary.Description,
				URL:         summary.ContentUrls.Desktop.Page,
				Image: WikiSummaryImage{
					URL:    summary.Thumbnail.Source,
					Width:  summary.Thumbnail.Width,
					Height: summary.Thumbnail.Height,
				},
			},
		},
	)
	return summaries, nil
}
