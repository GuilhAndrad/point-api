package repository

import (
	"context"
	"time"

	"github.com/GuilhAndrad/point-api/internal/model"
	"github.com/jmoiron/sqlx"
)

type RecordRepository interface {
    Create(ctx context.Context, record *model.TimeRecord) error
    FindByUserID(ctx context.Context, userID int64, from, to time.Time) ([]model.TimeRecord, error)
    FindLastByUserID(ctx context.Context, userID int64) (*model.TimeRecord, error)
    FindAll(ctx context.Context, from, to time.Time) ([]model.TimeRecord, error)
}

type recordRepository struct {
    db *sqlx.DB
}

func NewRecordRepository(db *sqlx.DB) RecordRepository {
    return &recordRepository{db: db}
}

func (r *recordRepository) Create(ctx context.Context, record *model.TimeRecord) error {
    query := `
        INSERT INTO time_records (user_id, type, timestamp)
        VALUES ($1, $2, $3)
        RETURNING id, created_at
    `
    return r.db.QueryRowContext(ctx, query,
        record.UserID, record.Type, record.Timestamp).
        Scan(&record.ID, &record.CreatedAt)
}

// FindLastByUserID busca o último registro do funcionário
// (usado para saber se o próximo deve ser entrada ou saída)
func (r *recordRepository) FindLastByUserID(ctx context.Context, userID int64) (*model.TimeRecord, error) {
    var record model.TimeRecord
    err := r.db.GetContext(ctx, &record, `
        SELECT * FROM time_records
        WHERE user_id = $1
        ORDER BY timestamp DESC
        LIMIT 1
    `, userID)
    if err != nil {
        return nil, err
    }
    return &record, nil
}

func (r *recordRepository) FindByUserID(ctx context.Context, userID int64, from, to time.Time) ([]model.TimeRecord, error) {
    var records []model.TimeRecord
    err := r.db.SelectContext(ctx, &records, `
        SELECT * FROM time_records
        WHERE user_id = $1 AND timestamp BETWEEN $2 AND $3
        ORDER BY timestamp ASC
    `, userID, from, to)
    return records, err
}

func (r *recordRepository) FindAll(ctx context.Context, from, to time.Time) ([]model.TimeRecord, error) {
    var records []model.TimeRecord
    err := r.db.SelectContext(ctx, &records, `
        SELECT * FROM time_records
        WHERE timestamp BETWEEN $1 AND $2
        ORDER BY user_id, timestamp ASC
    `, from, to)
    return records, err
}