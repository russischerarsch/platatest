package repository

import (
	"PlataTest/internal/domain"
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
)

const UniqueViolation = "23505"

type Repository struct {
	pool *pgxpool.Pool
}

func NewRepository(pool *pgxpool.Pool) *Repository {
	return &Repository{
		pool: pool,
	}
}
func (r *Repository) CreateQuote(ctx context.Context, pair string) (int64, error) {
	var id int64
	query := `
	INSERT INTO quotes (pair, created_at)
	VALUES ($1, $2)
	RETURNING id
	`
	err := r.pool.QueryRow(ctx, query, pair, time.Now()).Scan(&id)
	var pgError *pgconn.PgError
	if err != nil {
		if errors.As(err, &pgError) && pgError.Code == UniqueViolation { // номер ошибки дубликата
			return 0, fmt.Errorf("create quote pair: %s:  %w", pair, domain.ErrAlreadyExists)
		}
		return id, fmt.Errorf("create quote: %w", err)
	}
	return id, nil
}
func (r *Repository) GetQuoteByPair(ctx context.Context, pair string) (*domain.Quote, error) {
	var quote domain.Quote
	query := `
	SELECT id, pair, price, updated_at, status FROM quotes
	WHERE pair = $1
	LIMIT 1
	`
	err := r.pool.QueryRow(ctx, query, pair).Scan(
		&quote.ID,
		&quote.Pair,
		&quote.Price,
		&quote.UpdatedAt,
		&quote.Status,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, fmt.Errorf("get quote for pair %s: %w", pair, domain.ErrNotFound)
		}

		return nil, fmt.Errorf("get quote pair %s: %w", pair, err)
	}
	return &quote, nil
}

func (r *Repository) GetQuoteByID(ctx context.Context, id int64) (*domain.Quote, error) {
	var quote domain.Quote
	query := `
	SELECT id, pair, price, updated_at, status FROM quotes
	WHERE id = $1
	`
	err := r.pool.QueryRow(ctx, query, id).Scan(
		&quote.ID,
		&quote.Pair,
		&quote.Price,
		&quote.UpdatedAt,
		&quote.Status)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, fmt.Errorf("get quote id %d: %w", id, domain.ErrNotFound)
		}
		return nil, fmt.Errorf("get quote by id: %w", err)
	}
	return &quote, nil
}
func (r *Repository) UpdateQuote(ctx context.Context, q *domain.Quote) error {

	query := `
	UPDATE quotes 
	SET price = $1, updated_at = $2, status = $3
	WHERE id = $4
	`
	res, err := r.pool.Exec(ctx, query, q.Price, q.UpdatedAt, q.Status, q.ID)
	if err != nil {
		return fmt.Errorf("%w:update quote: %w", domain.ErrStorage, err)
	}
	if res.RowsAffected() == 0 {
		return fmt.Errorf("quote %d not found: %w", q.ID, domain.ErrNotFound)
	}
	return nil
}
func (r *Repository) UpdateStatus(ctx context.Context, id int64, status string) error {
	query := `
	UPDATE quotes 
	SET status = $1
	WHERE id = $2
	`
	res, err := r.pool.Exec(ctx, query, status, id)
	if err != nil {
		return fmt.Errorf("update status: %w", err)
	}
	if res.RowsAffected() == 0 {
		return fmt.Errorf("quote %d not found: %w", id, domain.ErrNotFound)
	}
	return nil
}

func (r *Repository) GetOldestPending(ctx context.Context) (*domain.Quote, error) {
	query := `
	SELECT id, pair, attempts FROM quotes
	WHERE status = 'pending'
	ORDER BY created_at
	LIMIT 1
	`
	quote := domain.Quote{}
	err := r.pool.QueryRow(ctx, query).Scan(&quote.ID, &quote.Pair, &quote.Attempt)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, domain.ErrNotFound
		}
		return nil, fmt.Errorf("get oldest pending: %w", err)
	}
	return &quote, nil
}

func (r *Repository) IncrementAttempts(ctx context.Context, id int64) error {
	query := `
	UPDATE quotes
	SET attempts = attempts + 1
	WHERE id = $1
	`
	if _, err := r.pool.Exec(ctx, query, id); err != nil {
		return err
	}
	return nil
}
