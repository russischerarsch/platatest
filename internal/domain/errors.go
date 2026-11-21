package domain

import "errors"

var ErrNotFound = errors.New("quote not found")
var ErrAPINotAvailable = errors.New("failed to connect third-party service")
var ErrInvalidPairFormat = errors.New("invalid currency pair format. Use format USDGBP")
var ErrAlreadyExists = errors.New("pair already exists")
var ErrStorage = errors.New("storage error")
var ErrDeniedAPI = errors.New("api denied access to conent")
