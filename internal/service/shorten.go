package service

import "context"

type URLStorage interface {
	StoreURL(ctx context.Context, code, url string) error
	GetURL(ctx context.Context, code string) (string, error)
}
