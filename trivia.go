package main

import (
	"context"
	"embed"
	"encoding/json"
	"fmt"
	"math/rand"

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

func randomizeArticle(ctx context.Context) ([]WikiSummary, error) {
	subj := wikipediaArticleTitles[rand.Intn(len(wikipediaArticleTitles))]
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
