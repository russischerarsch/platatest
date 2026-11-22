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

func setupAPI(t *testing.T, status int, body string) *httptest.Server {
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
	return srv
}

func pendingQuote(attempt int) *domain.Quote {
	return &domain.Quote{ID: 1, Pair: "eurusd", Attempt: attempt}
}

func TestUpdatePrice_success(t *testing.T) {
	setupAPI(t, http.StatusOK, `{"success":true,"rates":{"USD":1.1}}`)
	m := mocks.NewWorkerStore(t)
	var saved *domain.Quote

	m.On("GetOldestPending", mock.Anything).Return(pendingQuote(0), nil)
	m.On("UpdateQuote", mock.Anything, mock.Anything).Run(func(args mock.Arguments) {
		saved = args.Get(1).(*domain.Quote)
	}).Return(nil)

	updatePrices(t.Context(), m, http.DefaultClient, "key")

	require.NotNil(t, saved)
	assert.Equal(t, int64(1), saved.ID)
	assert.Equal(t, "completed", saved.Status)
	assert.Equal(t, 1.1, *saved.Price)
}
