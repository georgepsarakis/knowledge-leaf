package main

import (
	"context"
	"fmt"
	"maps"
	"slices"
	"strings"

	"go.uber.org/zap"

	"knowledgeleaf/app"
	"knowledgeleaf/externalapi/wikipedia"
)

func main() {
	application, cleanup, err := app.New()
	if err != nil {
		panic(err)
	}
	defer func() {
		if err := cleanup(); err != nil {
			panic(err)
		}
	}()

	ctx, cancel := context.WithTimeout(context.Background(), application.Cfg.ScheduledLoaderTimeout)
	defer cancel()
	application.Logger.Info("fetching data from wikipedia")
	scanner, onComplete, err := wikipedia.DownloadArticleDump(ctx)
	if err != nil {
		application.Logger.Fatal("wikipedia request failed", zap.Error(err))
	}

	articleTitles := make(map[string]struct{}, 1_000_000)
	var index int
	for scanner.Scan() {
		title, ok := normalizeTitle(scanner.Text())
		if !ok {
			continue
		}
		index++
		// Skip header row
		if index == 1 {
			continue
		}
		articleTitles[title] = struct{}{}
	}
	if err := scanner.Err(); err != nil {
		application.Logger.Fatal("error reading file", zap.Error(err))
	}
	if err := onComplete(); err != nil {
		application.Logger.Fatal("error completing Wikipedia dump request", zap.Error(err))
	}
	application.Logger.Info(
		"wikipedia article dump retrieval completed",
		zap.Int("total_titles", len(articleTitles)))

	allTitles := slices.Collect(maps.Keys(articleTitles))
	for batch := range slices.Chunk(allTitles, 1000) {
		index += len(batch)
		if err := application.Repository.BulkCreate(ctx, batch); err != nil {
			application.Logger.Fatal("persisting batch failed", zap.Error(err))
		}
		application.Logger.Info(fmt.Sprintf("created %d entries", index))

	}
}

func normalizeTitle(s string) (string, bool) {
	s = strings.TrimSpace(s)
	s = strings.ReplaceAll(s, `"`, "")
	isInvalid := len(s) > 30 || len(s) < 3 || strings.HasPrefix(s, "!")
	return s, !isInvalid
}
