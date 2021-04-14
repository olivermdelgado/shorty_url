package model

import "time"

type ShortURL struct {
	ID          uint
	UUID        string
	OriginalURL string
	ShortURL    string
	CreatedAt   time.Time
	UpdatedAt   time.Time
}

type ShortURLMetrics struct {
	ID        uint
	UUID      string
	ShortURL  string
	IPAddress string
	UserAgent string
	Count     int
	CreatedAt time.Time
	UpdatedAt time.Time
}
