package worker

import (
	"PlataTest/internal/domain"
	"PlataTest/internal/worker/mocks"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func TestUpdatePrice_success(t *testing.T) {
	setupAPI(t, http.StatusOK, `{"success":true,"rates":{"USD":1.1}}`)
	store := mocks.NewWorkerStore(t)
	var saved *domain.Quote

	store.On("GetOldestPending", mock.Anything).Return(pendingQuote(0), nil)
	store.On("UpdateQuote", mock.Anything, mock.Anything).Run(func(args mock.Arguments) {
		saved = args.Get(1).(*domain.Quote)
	}).Return(nil)

	updatePrices(t.Context(), store, http.DefaultClient, "key")

	require.NotNil(t, saved)
	assert.Equal(t, int64(1), saved.ID)
	assert.Equal(t, "completed", saved.Status)
	assert.Equal(t, 1.1, *saved.Price)
}

func TestUpdatePrice_tooManyAttempts(t *testing.T) {
	store := mocks.NewWorkerStore(t)
	store.On("GetOldestPending", mock.Anything).Return(pendingQuote(3), nil)
	store.On("UpdateStatus", mock.Anything, int64(1), failedSt).Return(nil)
	updatePrices(t.Context(), store, http.DefaultClient, "key")
}

func TestUpdatePrice_apiBadRequest(t *testing.T) {
	setupAPI(t, http.StatusBadRequest, "")
	store := mocks.NewWorkerStore(t)
	store.On("GetOldestPending", mock.Anything).Return(pendingQuote(1), nil)
	store.On("UpdateStatus", mock.Anything, int64(1), failedSt).Return(nil)
	store.On("IncrementAttempts", mock.Anything, mock.Anything).Return(nil)
	updatePrices(t.Context(), store, http.DefaultClient, "key")
}

func TestUpdateQuote_unauthorized(t *testing.T) {
	setupAPI(t, http.StatusUnauthorized, "")
	store := mocks.NewWorkerStore(t)
	err := updateQuote(t.Context(), store, http.DefaultClient, "", pendingQuote(1))

	require.ErrorIs(t, err, domain.ErrDeniedAPI)

}

func TestUpdatePrice_storageErr(t *testing.T) {
	setupAPI(t, http.StatusOK, `{"success":true,"rates":{"USD":1.1}}`)
	store := mocks.NewWorkerStore(t)
	store.On("GetOldestPending", mock.Anything).Return(pendingQuote(1), nil)
	store.On("UpdateQuote", mock.Anything, mock.Anything).Return(domain.ErrStorage)
	updatePrices(t.Context(), store, http.DefaultClient, "key")
}

func TestUpdateQuote_apiNotAvailable(t *testing.T) {
	old := apiBaseURL
	apiBaseURL = "http://127.0.0.1:1"
	defer func() { apiBaseURL = old }()
	store := mocks.NewWorkerStore(t)

	err := updateQuote(t.Context(), store, http.DefaultClient, "key", pendingQuote(1))

	require.ErrorIs(t, err, domain.ErrAPINotAvailable)
}

func setupAPI(t *testing.T, status int, body string) {
	t.Helper()
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(status)
		w.Write([]byte(body))
	}))
	old := apiBaseURL
	apiBaseURL = srv.URL
	t.Cleanup(func() {
		srv.Close()
		apiBaseURL = old
	})
}

func pendingQuote(attempt int) *domain.Quote {
	return &domain.Quote{ID: 1, Pair: "eurusd", Attempt: attempt}
}
