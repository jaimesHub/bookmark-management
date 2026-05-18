package service

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/jaimesHub/bookmark-management/pkg/stringutils"
	"github.com/redis/go-redis/v9"
)

// ErrCodeNotFound is returned when a shortened code does not exist in storage.
// Handlers compare via errors.Is to map this to HTTP 404 without importing the storage package.
var ErrCodeNotFound = errors.New("code not found")

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
	GetOriginalURL(ctx context.Context, code string) (string, error)
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

// GetOriginalURL resolves a shortened code back to its original URL.
// Returns ErrCodeNotFound when the code does not exist; other errors are wrapped with context.
func (s *shortenService) GetOriginalURL(ctx context.Context, code string) (string, error) {
	url, err := s.repo.GetURL(ctx, code)
	if err != nil {
		if errors.Is(err, redis.Nil) {
			return "", ErrCodeNotFound
		}
		return "", fmt.Errorf("get original url: %w", err)
	}
	return url, nil
}
