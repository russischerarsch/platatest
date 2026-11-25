package transport_test

import (
	"PlataTest/internal/domain"
	"PlataTest/internal/transport"
	"PlataTest/internal/transport/mocks"
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func TestCreateQuote_success(t *testing.T) {
	reqBody := struct {
		Pair string `json:"pair"`
	}{
		Pair: "EURUSD",
	}
	data, err := json.Marshal(reqBody)
	require.NoError(t, err)
	req := httptest.NewRequest(
		http.MethodPost,
		"/api/quotes",
		bytes.NewBuffer(data),
	)
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	mux := http.NewServeMux()
	svc := mocks.NewService(t)
	svc.On("CreateQuote", mock.Anything, "EURUSD").Return(int64(1), nil)
	h := transport.NewHandler(svc)
	mux.HandleFunc("POST /api/quotes", h.CreateQuote)

	mux.ServeHTTP(rec, req)
	require.Equal(t, http.StatusCreated, rec.Code)
	require.Equal(t, "application/json", rec.Header().Get("Content-Type"))
	var id transport.СreateQuoteResponse
	err = json.NewDecoder(rec.Body).Decode(&id)
	require.NoError(t, err)
	assert.Equal(t, int64(1), id.ID)
}

func TestGetQuoteByID_success(t *testing.T) {
	req := httptest.NewRequest(
		http.MethodGet,
		"/api/quotes/1",
		nil,
	)
	rec := httptest.NewRecorder()

	svc := mocks.NewService(t)
	svc.On("GetQuoteByID", mock.Anything, int64(1)).Return(&domain.Quote{Pair: "EURUSD"}, nil)

	h := transport.NewHandler(svc)

	mux := http.NewServeMux()

	mux.HandleFunc("GET /api/quotes/{id}", h.GetByID)

	mux.ServeHTTP(rec, req)
	require.Equal(t, http.StatusOK, rec.Code)

	var resp struct {
		Pair string `json:"pair"`
	}
	err := json.NewDecoder(rec.Body).Decode(&resp)
	require.NoError(t, err)
	assert.Equal(t, "EURUSD", resp.Pair)
}
func TestGetLatestQuote_success(t *testing.T) {
	req := httptest.NewRequest(
		http.MethodGet,
		"/api/quotes/latest?pair=EURUSD",
		nil,
	)
	rec := httptest.NewRecorder()

	svc := mocks.NewService(t)
	svc.On("GetQuoteByPair", mock.Anything, "EURUSD").Return(&domain.Quote{Pair: "EURUSD"}, nil)

	h := transport.NewHandler(svc)

	mux := http.NewServeMux()

	mux.HandleFunc("GET /api/quotes/latest", h.GetQuoteByPair)

	mux.ServeHTTP(rec, req)
	require.Equal(t, http.StatusOK, rec.Code)

	var resp struct {
		Pair string `json:"pair"`
	}
	err := json.NewDecoder(rec.Body).Decode(&resp)
	require.NoError(t, err)
	assert.Equal(t, "EURUSD", resp.Pair)
}
