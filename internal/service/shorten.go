package service

import (
	"context"
	"time"

	"github.com/jaimesHub/bookmark-management/pkg/stringutils"
)

// example: https://raviatluri.in/articles/using-mockery-go-generate
//
//go:generate mockery --name URLStorage --filename urlstorage.go
type URLStorage interface {
	StoreURL(ctx context.Context, code, url string, exp time.Duration) error
	GetURL(ctx context.Context, code string) (string, error)
}

const (
	// urlCodeLength is the fixed length of a generated short URL code.
	// 7 alphanumeric chars → 62^7 ≈ 3.5 trillion combinations, sufficient to avoid collisions.
	urlCodeLength = 7
)

//go:generate mockery --name ShortenService --filename shorten_service.go
type ShortenService interface {
	ShortenURL(ctx context.Context, url string, exp time.Duration) (string, error)
}

type shortenService struct {
	repo URLStorage
}

func NewShortenService(repo URLStorage) ShortenService {
	return &shortenService{repo: repo}
}

func (s *shortenService) ShortenURL(ctx context.Context, url string, exp time.Duration) (string, error) {
	// create key
	urlCode, err := stringutils.GenerateCode(urlCodeLength)
	if err != nil {
		return "", err
	}

	// adding to storage (redis)
	err = s.repo.StoreURL(ctx, urlCode, url, exp)
	if err != nil {
		return "", err
	}

	// return key
	return urlCode, nil
}
