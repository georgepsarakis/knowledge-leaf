package main

import (
	"bufio"
	"compress/gzip"
	"context"
	"fmt"
	"net/http"
	"strings"
	"time"

	"knowledgeleaf/app"
)

const wikipediaDumpURL = "https://dumps.wikimedia.org/enwiki/latest/enwiki-latest-all-titles-in-ns0.gz"

func main() {
	application, cleanup, err := app.New()
	if err != nil {
		panic(err)
	}
	defer cleanup()

	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Minute)
	defer cancel()
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, wikipediaDumpURL, nil)
	if err != nil {
		panic(err)
	}
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		panic(err)
	}
	defer resp.Body.Close()
	gz, err := gzip.NewReader(resp.Body)
	if err != nil {
		panic(err)
	}
	scanner := bufio.NewScanner(gz)
	index := 0
	var batch []string
	batchSize := 1000
	for scanner.Scan() {
		title := strings.TrimSpace(scanner.Text())
		if len(title) > 30 || len(title) < 3 || strings.HasPrefix(title, "!") {
			continue
		}
		batch = append(batch, title)
		index++
		if len(batch) >= batchSize {
			if err := persist(ctx, application, batch, index); err != nil {
				panic(err)
			}
			batch = batch[:0]
		}

		if index%1000 == 0 {
			application.Logger.Info(fmt.Sprintf("created %d entries", index))
		}
	}
	if len(batch) > 0 {
		if err := persist(ctx, application, batch, index); err != nil {
			panic(err)
		}
	}

	if err := scanner.Err(); err != nil {
		fmt.Println("Error reading file:", err)
	}
}

func persist(ctx context.Context, app app.App, batch []string, indexOffset int) error {
	if len(batch) == 0 {
		return nil
	}
	members := make([]any, 0, len(batch))
	for _, t := range batch {
		members = append(members, t)
	}
	if err := app.RedisClient.SAdd(ctx, "datasource:wikipedia", members...).Err(); err != nil {
		return err
	}
	return nil
}
