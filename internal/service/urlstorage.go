package service

// DONE: moved to internal/service/shorten.go
// TODO: DELETE THIS FILE IF NOT USED ANYMORE

// import (
// 	"context"

// 	"github.com/jaimesHub/bookmark-management/internal/repository"
// 	"github.com/jaimesHub/bookmark-management/pkg/stringutils"
// )

// const (
// 	urlCodeLength = 7
// )

// type ShortenUrl interface {
// 	ShortenUrl(ctx context.Context, url string) (string, error)
// }

// type shortenUrl struct {
// 	repo repository.UrlStorage
// }

// func NewShortenUrl(repo repository.UrlStorage) ShortenUrl {
// 	return &shortenUrl{repo: repo}
// }

// func (s *shortenUrl) ShortenUrl(ctx context.Context, url string) (string, error) {
// 	// create key
// 	urlCode, err := stringutils.GenerateCode(urlCodeLength)
// 	if err != nil {
// 		return "", err
// 	}

// 	// adding to storage (redis)
// 	err = s.repo.StoreURL(ctx, urlCode, url) // ?
// 	if err != nil {
// 		return "", err
// 	}

// 	// return key
// 	return urlCode, nil
// }
