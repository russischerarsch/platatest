package dbconnection

import (
	"context"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

const (
	maxAttempts = 10
	retryDelay  = 2 * time.Second
	pingTimeout = 2 * time.Second
)

func NewPool(ctx context.Context, dsn string) (*pgxpool.Pool, error) {
	pool, err := pgxpool.New(ctx, dsn)
	if err != nil {
		return nil, fmt.Errorf("create a pool, err %w", err)
	}
	for attempt := 0; attempt < maxAttempts; attempt++ {

		delay := retryDelay
		if attempt == 0 {
			delay = 0 * time.Second
		}
		select {
		case <-ctx.Done():
			pool.Close()
			return nil, ctx.Err()
		case <-time.After(delay):
			childCtx, cancel := context.WithTimeout(ctx, pingTimeout)
			if err = pool.Ping(childCtx); err == nil {
				cancel()
				return pool, nil
			}
			cancel()

		}
	}
	pool.Close()
	return nil, fmt.Errorf("connect to postgres after %d attempts, err %w", maxAttempts, err)
}
