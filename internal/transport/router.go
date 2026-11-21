package transport

import (
	"net/http"
)

func SetupRouter(handlers *Handlers) http.Handler {

	mux := http.NewServeMux()

	mux.HandleFunc("POST /api/quotes", handlers.CreateQuote)
	mux.HandleFunc("GET /api/quotes/latest", handlers.GetQuoteByPair)
	mux.HandleFunc("GET /api/quotes/{id}", handlers.GetByID)

	return mux
}
