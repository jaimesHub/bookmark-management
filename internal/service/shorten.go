package service

import (
	"context"
	"fmt"
	"time"

	"github.com/jaimesHub/bookmark-management/pkg/stringutils"
)

// example: https://raviatluri.in/articles/using-mockery-go-generate
//
//go:generate mockery --name URLStorage --filename urlstorage.go
type URLStorage interface {
	StoreURL(ctx context.Context, code, url string, exp time.Duration) error
	GetURL(ctx context.Context, code string) (string, error)
	StoreURLIfNotExists(ctx context.Context, code, url string, exp time.Duration) (bool, error)
}

const (
	// urlCodeLength is the fixed length of a generated short URL code.
	// 7 alphanumeric chars → 62^7 ≈ 3.5 trillion combinations, sufficient to avoid collisions.
	urlCodeLength = 7
	maxRetries    = 3
)

//go:generate mockery --name ShortenService --filename shorten_service.go

// ShortenService is the business-logic contract for URL shortening.
// Callers supply the original URL and desired TTL; the service returns an opaque short code.
type ShortenService interface {
	ShortenURL(ctx context.Context, url string, exp time.Duration) (string, error)
}

type shortenService struct {
	repo URLStorage
}

// NewShortenService constructs a ShortenService backed by the given URLStorage.
func NewShortenService(repo URLStorage) ShortenService {
	return &shortenService{repo: repo}
}

// ShortenURL generates a unique short code for url, stores it with the given TTL, and returns the code.
// It retries up to maxRetries times if the generated code already exists in the store.
func (s *shortenService) ShortenURL(ctx context.Context, url string, exp time.Duration) (string, error) {
	for range maxRetries {
		code, err := stringutils.GenerateCode(urlCodeLength)
		if err != nil {
			return "", err
		}

		stored, err := s.repo.StoreURLIfNotExists(ctx, code, url, exp)
		if err != nil {
			return "", err
		}
		if stored {
			return code, nil
		}
	}
	return "", fmt.Errorf("failed to generate unique code after %d attempts", maxRetries)
}
