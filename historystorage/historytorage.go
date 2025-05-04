package historystorage

import (
	"context"
	"fmt"
	"io"
	"os"

	"gitverse.ru/udonetsm/aiserver/configs"
	"gitverse.ru/udonetsm/aiserver/history"
	"gitverse.ru/udonetsm/aiserver/logger"
)

type historystorage struct {
	historyStorageConfig configs.HistoryStorageConfig
	logger               logger.Logger
	storage              io.WriteCloser
}

type HistoryStorage interface {
	Save(ctx context.Context, history history.History) error
	Configure(ctx context.Context) error
	CloseStorage() error
}

func (h *historystorage) Configure(ctx context.Context) error {
	err := h.createWriteCloser(ctx)
	if err != nil {
		return fmt.Errorf("%w", err)
	}
	return nil
}

func (h *historystorage) Save(ctx context.Context, history history.History) error {
	defer history.Remove()
	defer h.storage.Close()
	_, err := fmt.Fprintln(h.storage, history)
	if err != nil {
		return fmt.Errorf("historyStorage error: %w", err)
	}
	return nil
}

func (hs *historystorage) createWriteCloser(ctx context.Context) error {
	hs.logger.Infof("creating %s... ", hs.historyStorageConfig.HistorySource())
	file, err := os.OpenFile(hs.historyStorageConfig.HistorySource(), os.O_WRONLY|os.O_CREATE, 0755)
	for {
		select {
		case <-ctx.Done():
			hs.logger.Info(ctx.Err())
			return fmt.Errorf("%w", ctx.Err())
		default:
			if err != nil {
				return fmt.Errorf("create writecloser error: %w", err)
			}
			hs.storage = file
			return nil
		}
	}
}

func (hs *historystorage) CloseStorage() error {
	err := hs.storage.Close()
	if err != nil {
		return fmt.Errorf("%w", err)
	}
	return nil
}

func NewHistoryStorage(logger logger.Logger, historyStorageConfig configs.HistoryStorageConfig) HistoryStorage {
	return &historystorage{logger: logger, historyStorageConfig: historyStorageConfig}
}
