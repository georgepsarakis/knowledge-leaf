package main

import (
	"bufio"
	"compress/gzip"
	"context"
	"fmt"
	"strings"

	"github.com/georgepsarakis/go-httpclient"
	"go.uber.org/zap"

	"knowledgeleaf/app"
)

const wikipediaDumpURL = "https://dumps.wikimedia.org/enwiki/latest/enwiki-latest-all-titles-in-ns0.gz"

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
	httpClient := httpclient.New()
	resp, err := httpClient.Get(ctx, wikipediaDumpURL)
	if err != nil {
		application.Logger.Fatal("wikipedia request failed", zap.Error(err))
	}
	defer resp.Body.Close()
	gz, err := gzip.NewReader(resp.Body)
	if err != nil {
		application.Logger.Fatal("wikipedia request failed", zap.Error(err))
	}
	scanner := bufio.NewScanner(gz)

	articleTitles := make(map[string]struct{}, 1_000_000)
	var index int
	for scanner.Scan() {
		title := normalizeTitle(scanner.Text())
		if len(title) > 30 || len(title) < 3 || strings.HasPrefix(title, "!") {
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
		application.Logger.Fatal("error reading file:", zap.Error(err))
	}
	var batch []string
	batchSize := 1000
	index = 0

	for t := range articleTitles {
		batch = append(batch, t)
		index++
		if len(batch) >= batchSize {
			if err := application.Repository.BulkCreate(ctx, batch); err != nil {
				application.Logger.Fatal("persisting batch failed", zap.Error(err))
			}
			batch = batch[:0]
		}

		if index%1000 == 0 {
			application.Logger.Info(fmt.Sprintf("created %d entries", index))
		}
	}
	if err := application.Repository.BulkCreate(ctx, batch); err != nil {
		application.Logger.Fatal("persisting batch failed", zap.Error(err))
	}
}

func normalizeTitle(s string) string {
	s = strings.TrimSpace(s)
	s = strings.ReplaceAll(s, `"`, "")
	return s
}
