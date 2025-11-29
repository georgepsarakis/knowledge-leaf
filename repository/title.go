package repository

import (
	"context"
	"errors"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgerrcode"
	"github.com/jackc/pgx/v5/pgconn"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"

	"knowledgeleaf/database"
)

type Repository interface {
	BulkCreate(context.Context, []string) error
	CurrentBucketValue(context.Context) (int64, error)
	NextBucketValue(context.Context) (int64, error)
}

type Title struct {
	ID        string
	Title     string
	NumericID int
	CreatedAt time.Time
	UpdatedAt time.Time
}

type postgresRepository struct {
	db *gorm.DB
}

const sequenceNameBucketID = "wikipedia_titles_numeric_id_seq"

func (p postgresRepository) BulkCreate(ctx context.Context, titles []string) error {
	if len(titles) == 0 {
		return nil
	}

	var existingCount int64
	existingArticleCountQuery := p.db.WithContext(ctx).Model(&database.WikipediaTitle{}).
		Where("title IN ?", titles)
	if err := existingArticleCountQuery.Count(&existingCount).Error; err != nil {
		return err
	}
	diff := len(titles) - int(existingCount)
	if diff <= 0 {
		return nil
	}

	var bucketID int64
	if diff <= 50 {
		b, err := p.CurrentBucketValue(ctx)
		if err != nil {
			var pgErr *pgconn.PgError
			if !errors.As(err, &pgErr) || pgErr.Code != pgerrcode.ObjectNotInPrerequisiteState {
				return err
			}
		} else {
			bucketID = b
		}
	}
	if bucketID == 0 {
		b, err := p.NextBucketValue(ctx)
		if err != nil {
			return err
		}
		bucketID = b
	}
	rows := make([]*database.WikipediaTitle, 0, len(titles))
	for _, item := range titles {
		rows = append(rows, &database.WikipediaTitle{
			ID:        uuid.NewString(),
			Title:     item,
			NumericID: int(bucketID),
		})
	}

	tx := p.db.WithContext(ctx).
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

func (p postgresRepository) CurrentBucketValue(ctx context.Context) (int64, error) {
	var n int64
	err := p.db.WithContext(ctx).Raw("SELECT CURRVAL(?)", sequenceNameBucketID).Scan(&n).Error
	return n, err
}

func (p postgresRepository) NextBucketValue(ctx context.Context) (int64, error) {
	var n int64
	err := p.db.WithContext(ctx).Raw("SELECT NEXTVAL(?)", sequenceNameBucketID).Scan(&n).Error
	return n, err
}

func NewPostgresRepository(db *gorm.DB) Repository {
	return postgresRepository{db: db}
}
