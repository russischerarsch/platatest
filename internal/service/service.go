package service

import (
	"PlataTest/internal/domain"
	"context"
	"fmt"
	"strings"
)

type Repository interface {
	CreateQuote(ctx context.Context, pair string) (int64, error)
	GetQuoteByPair(ctx context.Context, pair string) (*domain.Quote, error)
	GetQuoteByID(ctx context.Context, ID int64) (*domain.Quote, error)
	UpdateQuote(ctx context.Context, q *domain.Quote) error
	UpdateStatus(ctx context.Context, ID int64, status string) error
	GetOldestPending(ctx context.Context) (*domain.Quote, error)
	IncrementAttempts(ctx context.Context, id int64) error
}
type QuoteService struct {
	repo Repository
}

func NewService(repo Repository) *QuoteService {
	return &QuoteService{
		repo: repo}
}
func (q *QuoteService) CreateQuote(ctx context.Context, pair string) (int64, error) {
	pair = normalizePair(pair)
	if err := validatePair(pair); err != nil {
		return 0, err
	}

	id, err := q.repo.CreateQuote(ctx, pair)
	if err != nil {
		return 0, err
	}
	return id, nil
}

func (q *QuoteService) GetQuoteByID(ctx context.Context, id int64) (*domain.Quote, error) {
	quote, err := q.repo.GetQuoteByID(ctx, id)
	if err != nil {
		return nil, err
	}
	quote.Pair = strings.ToUpper(quote.Pair)
	return quote, nil
}

func (q *QuoteService) GetQuoteByPair(ctx context.Context, pair string) (*domain.Quote, error) {
	pair = normalizePair(pair)
	if err := validatePair(pair); err != nil {
		return nil, err
	}
	quote, err := q.repo.GetQuoteByPair(ctx, pair)
	if err != nil {
		return nil, err
	}
	quote.Pair = strings.ToUpper(quote.Pair)
	return quote, nil
}
func validatePair(pair string) error {
	if len(pair) != 6 {
		return fmt.Errorf("pair must have 6 letters: %w", domain.ErrInvalidPairFormat)
	}

	for _, r := range pair {
		if r < 'a' || r > 'z' {
			return fmt.Errorf("pair must contain only letters: %w", domain.ErrInvalidPairFormat)
		}
	}

	return nil
}

func normalizePair(pair string) string {
	pair = strings.TrimSpace(pair)
	pair = strings.ToLower(pair)
	return pair
}
