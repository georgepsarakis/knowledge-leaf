package main

import (
	"bufio"
	"compress/gzip"
	"context"
	"errors"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgerrcode"
	"github.com/jackc/pgx/v5/pgconn"
	"go.uber.org/zap"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"

	"knowledgeleaf/app"
	"knowledgeleaf/database"
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

	normalize := func(s string) string {
		s = strings.TrimSpace(s)
		s = strings.ReplaceAll(s, `"`, "")
		return s
	}

	articleTitles := make(map[string]struct{}, 1_000_000)
	var index int
	for scanner.Scan() {
		title := normalize(scanner.Text())
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
		application.Logger.Error("error reading file:", zap.Error(err))
	}
	var batch []string
	batchSize := 1000
	index = 0

	for t := range articleTitles {
		batch = append(batch, t)
		index++
		if len(batch) >= batchSize {
			if err := persist(ctx, application, batch); err != nil {
				panic(err)
			}
			batch = batch[:0]
		}

		if index%1000 == 0 {
			application.Logger.Info(fmt.Sprintf("created %d entries", index))
		}
	}
	if len(batch) > 0 {
		if err := persist(ctx, application, batch); err != nil {
			application.Logger.Error("error persisting entries", zap.Error(err))
		}
	}

}

func persist(ctx context.Context, app app.App, batch []string) error {
	if len(batch) == 0 {
		return nil
	}

	var existingCount int64
	existingArticleCountQuery := app.PostgresConnection.Model(&database.WikipediaTitle{}).Where("title IN ?", batch)
	if err := existingArticleCountQuery.Count(&existingCount).Error; err != nil {
		return err
	}
	diff := len(batch) - int(existingCount)
	if diff <= 0 {
		return nil
	}

	var bucketID int64
	if diff <= 50 {
		b, err := currentSequenceValue(app.PostgresConnection.WithContext(ctx), "wikipedia_titles_numeric_id_seq")
		if err != nil {
			return err
		}
		bucketID = b
	} else {
		b, err := nextSequenceValue(app.PostgresConnection.WithContext(ctx), "wikipedia_titles_numeric_id_seq")
		if err != nil {
			return err
		}
		bucketID = b
	}
	rows := make([]*database.WikipediaTitle, 0, len(batch))
	for _, item := range batch {
		rows = append(rows, &database.WikipediaTitle{
			ID:        uuid.NewString(),
			Title:     item,
			NumericID: int(bucketID),
		})
	}

	tx := app.PostgresConnection.WithContext(ctx).
		Clauses(clause.OnConflict{DoNothing: true}).
		Create(rows)
	if err := tx.Error; err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) {
			if pgErr.Code == pgerrcode.UniqueViolation {
				return nil
			}
		}
		return tx.Error
	}
	return nil
}

func nextSequenceValue(db *gorm.DB, sequenceName string) (int64, error) {
	var n int64
	err := db.Raw("SELECT NEXTVAL(?)", sequenceName).Scan(&n).Error
	return n, err
}

func currentSequenceValue(db *gorm.DB, sequenceName string) (int64, error) {
	var n int64
	err := db.Raw("SELECT CURRVAL(?)", sequenceName).Scan(&n).Error
	return n, err
}
