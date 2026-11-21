package dto

import "time"

type Quote struct {
	Pair      string     `json:"pair"`
	Price     *float64   `json:"price"`
	UpdatedAt *time.Time `json:"updated_at"`
	Status    string     `json:"status"`
}
