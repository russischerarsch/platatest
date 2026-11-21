package domain

import "time"

type Quote struct {
	ID        int64
	Pair      string
	Price     *float64
	Attempt   int
	UpdatedAt *time.Time
	CreatedAt time.Time
	Status    string
}

type Info struct {
	Success bool               `json:"success"`
	Rates   map[string]float64 `json:"rates"`
	Error   *struct {
		Code string `json:"code"`
		Info string `json:"info"`
	} `json:"error"`
}
