package worker

import (
	"PlataTest/internal/domain"
	"PlataTest/internal/service"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/http"
	"time"
)

const (
	requestTimeout  = 20 * time.Second
	tickerInterval  = 10 * time.Second
	dbTimeout       = 5 * time.Second
	maxAttempts     = 3
	validPairLength = 6
	failedSt        = "failed"
)

func RunPriceUpdater(
	ctx context.Context,
	repository service.Repository,
	client *http.Client,
	apiKey string,
) error {
	ticker := time.NewTicker(tickerInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-ticker.C:
			updatePrices(ctx, repository, client, apiKey)
		}
	}
}

func updatePrices(
	ctx context.Context,
	repository service.Repository,
	client *http.Client,
	apiKey string,
) {
	queryCtx, cancel := context.WithTimeout(ctx, dbTimeout)

	quote, err := repository.GetOldestPending(queryCtx)
	cancel()
	if err != nil {
		if errors.Is(err, domain.ErrNotFound) {
			return
		}
		log.Printf("get pending quote: %v", err)
		return
	}
	if quote.Attempt >= maxAttempts {
		statusCtx, cancel := context.WithTimeout(ctx, dbTimeout)
		if err := markFailed(statusCtx, repository, quote.ID); err != nil {
			log.Printf("mark %s failed status: %v", quote.Pair, err)
			cancel()
			return
		}
		log.Printf("quote %s has run out of attempts", quote.Pair)
		cancel()
		return
	}
	if err := updateQuote(ctx, repository, client, apiKey, quote); err != nil {
		if errors.Is(err, domain.ErrDeniedAPI) {
			log.Printf("api: %v", err)
			return
		}
		if !errors.Is(err, domain.ErrStorage) {
			incCtx, cancel := context.WithTimeout(ctx, dbTimeout)
			if updErr := repository.IncrementAttempts(incCtx, quote.ID); updErr != nil {
				log.Printf("increment attempt: %v", updErr)
			}
			cancel()
		}
		log.Printf("pair %s: %v", quote.Pair, err)
	}

}

func updateQuote(
	ctx context.Context,
	repository service.Repository,
	client *http.Client,
	apiKey string,
	q *domain.Quote,
) error {

	childCtx, cancel := context.WithTimeout(ctx, requestTimeout)
	defer cancel()
	if len(q.Pair) != validPairLength {
		if err := markFailed(childCtx, repository, q.ID); err != nil {
			log.Printf("update status: %v", err)
		}

		return fmt.Errorf("pair %s less 6", q.Pair)
	}

	URL := fmt.Sprintf(
		"http://api.exchangeratesapi.io/v1/latest?access_key=%s&base=%s&symbols=%s",
		apiKey, q.Pair[:3], q.Pair[3:],
	)

	req, err := http.NewRequestWithContext(childCtx, http.MethodGet, URL, nil)
	if err != nil {
		return fmt.Errorf("create req: %w", err)
	}
	r, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("%w: %w", domain.ErrAPINotAvailable, err)
	}
	defer r.Body.Close()
	if r.StatusCode != http.StatusOK {
		if r.StatusCode == http.StatusTooManyRequests || r.StatusCode == http.StatusUnauthorized {
			return domain.ErrDeniedAPI
		}
		if r.StatusCode == http.StatusBadRequest {
			updateCtx, cancel := context.WithTimeout(ctx, dbTimeout)
			if err := markFailed(updateCtx, repository, q.ID); err != nil {
				log.Printf("update status: %v", err)
			}
			cancel()
		}
		return fmt.Errorf("API service returned %d status code", r.StatusCode)
	}

	var response domain.Info
	if err := json.NewDecoder(r.Body).Decode(&response); err != nil {
		return fmt.Errorf("decode response: %w", err)
	}
	if !response.Success {
		if response.Error != nil {
			return fmt.Errorf(
				"API error %s: %s",
				response.Error.Code,
				response.Error.Info,
			)
		}
		return errors.New("API request failed")
	}

	if len(response.Rates) == 0 {
		return fmt.Errorf("empty rates")
	}

	now := time.Now()
	var rate domain.Quote
	for currency, value := range response.Rates {
		rate = domain.Quote{
			ID:        q.ID,
			Pair:      currency,
			Price:     &value,
			UpdatedAt: &now,
			Status:    "completed",
		}
	}
	childCtx, cancel = context.WithTimeout(ctx, dbTimeout)
	defer cancel()
	if err := repository.UpdateQuote(childCtx, &rate); err != nil {
		return fmt.Errorf("update quote: %w", err)
	}
	return nil
}

func markFailed(ctx context.Context, repository service.Repository, id int64) error {
	if err := repository.UpdateStatus(ctx, id, failedSt); err != nil {
		return err
	}
	return nil
}
