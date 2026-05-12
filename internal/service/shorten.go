package service

import (
	"context"
	"time"

	"github.com/jaimesHub/bookmark-management/pkg/stringutils"
)

type URLStorage interface {
	StoreURL(ctx context.Context, code, url string, exp time.Duration) error
	GetURL(ctx context.Context, code string) (string, error)
}

const (
	urlCodeLength = 7
)

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
