package service

import (
	"log/slog"
	"sync"
)

type Service struct {
	logger *slog.Logger
	wg     *sync.WaitGroup
}

func New(logger *slog.Logger, wg *sync.WaitGroup) Service {
	return Service{
		logger: logger,
		wg:     wg,
	}
}
