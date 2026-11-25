package transport

import (
	"PlataTest/internal/domain"
	dto "PlataTest/internal/transport/dto"

	"context"
	"encoding/json"
	"errors"
	"log"
	"net/http"
	"strconv"
)

type Service interface {
	CreateQuote(ctx context.Context, pair string) (int64, error)
	GetQuoteByID(ctx context.Context, id int64) (*domain.Quote, error)
	GetQuoteByPair(ctx context.Context, pair string) (*domain.Quote, error)
}

type Handlers struct {
	service Service
}
type CreateQuoteResponse struct {
	ID int64 `json:"id"`
}

func NewHandler(service Service) *Handlers {
	return &Handlers{service: service}
}

type errorResponse struct {
	Message string `json:"error"`
}

func (h *Handlers) CreateQuote(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Pair string `json:"pair"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSONError(w, http.StatusBadRequest, "invalid request body")
		return
	}
	id, err := h.service.CreateQuote(r.Context(), req.Pair)
	if err != nil {
		if errors.Is(err, domain.ErrInvalidPairFormat) {
			writeJSONError(w, http.StatusBadRequest, "invalid format")
			return
		} else if errors.Is(err, domain.ErrAlreadyExists) {
			writeJSONError(w, http.StatusConflict, "pair already exists")
			return
		}
		log.Printf("create quote: %v", err)
		writeJSONError(w, http.StatusInternalServerError, "internal error")
		return
	}

	writeJSON(w, http.StatusCreated, CreateQuoteResponse{ID: id})
}

func (h *Handlers) GetByID(w http.ResponseWriter, r *http.Request) {
	idStr := r.PathValue("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		writeJSONError(w, http.StatusBadRequest, "invalid id")
		return
	}
	quote, err := h.service.GetQuoteByID(r.Context(), int64(id))
	if err != nil {
		if errors.Is(err, domain.ErrNotFound) {
			writeJSONError(w, http.StatusNotFound, "quote not found")
			return
		}
		log.Printf("get quote by id %d: %v", id, err)
		writeJSONError(w, http.StatusInternalServerError, "internal error")
		return
	}
	quoteDTO := dto.Quote{
		Pair:      quote.Pair,
		Price:     quote.Price,
		UpdatedAt: quote.UpdatedAt,
		Status:    quote.Status,
	}
	writeJSON(w, http.StatusOK, quoteDTO)
}

func (h *Handlers) GetQuoteByPair(w http.ResponseWriter, r *http.Request) {
	pair := r.URL.Query().Get("pair")
	if pair == "" {
		writeJSONError(w, http.StatusBadRequest, "pair is empty")
		return
	}
	quote, err := h.service.GetQuoteByPair(r.Context(), pair)
	if err != nil {
		if errors.Is(err, domain.ErrNotFound) {
			writeJSONError(w, http.StatusNotFound, "quote not found")
			return
		} else if errors.Is(err, domain.ErrInvalidPairFormat) {
			writeJSONError(w, http.StatusBadRequest, "invalid format")
			return
		}
		log.Printf("find quote by pair %s: %v", pair, err)
		writeJSONError(w, http.StatusInternalServerError, "internal error")
		return
	}
	quoteDTO := dto.Quote{
		Pair:      quote.Pair,
		Price:     quote.Price,
		UpdatedAt: quote.UpdatedAt,
		Status:    quote.Status,
	}
	writeJSON(w, http.StatusOK, quoteDTO)
}

func writeJSON(w http.ResponseWriter, status int, v interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	if err := json.NewEncoder(w).Encode(v); err != nil {
		log.Printf("write response: %v", err)
	}
}

func writeJSONError(w http.ResponseWriter, status int, msg string) {
	writeJSON(w, status, errorResponse{Message: msg})
}
