package configs

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"
)

type historyStorageConfig struct {
	hitorysource string
	apiKey       string
}

type HistoryStorageConfig interface {
	HistorySource() string
	Configure(source string) error
}

func (h *historyStorageConfig) HistorySource() string {
	return h.hitorysource
}

func (h *historyStorageConfig) home() (string, error) {
	home := os.Getenv("HOME")
	if home == "" {
		return "", fmt.Errorf("not found home directory")
	}
	return home, nil
}

func (h *historyStorageConfig) createHistoryFolderfNotExists(parent string) error {
	_, err := os.Stat(parent)
	if os.IsNotExist(err) {
		err := os.Mkdir(parent, 0755)
		if err != nil {
			return fmt.Errorf("%s create chat history folder fail: %w", parent, err)
		}
	}
	return nil
}

func (h *historyStorageConfig) analizeSource(source string) error {
	if source == "" {
		home, err := h.home()
		if err != nil {
			return fmt.Errorf("historyStorage.Configure() error: %w", err)
		}
		source = filepath.Join(home, "chathistory")
	}

	splited := strings.Split(source, string(filepath.Separator))
	splited[0] = string(filepath.Separator)
	for _, dir := range splited {
		h.hitorysource = filepath.Join(h.hitorysource, dir)
		err := h.createHistoryFolderfNotExists(h.hitorysource)
		if err != nil {
			return fmt.Errorf("%w", err)
		}
	}
	h.hitorysource = filepath.Join(h.hitorysource, h.apiKey, time.Now().Format("Jan-2-2006_15:04:05"))
	return nil
}

func (h *historyStorageConfig) Configure(source string) error {
	err := h.analizeSource(source)
	if err != nil {
		return fmt.Errorf("%w", err)
	}
	return nil
}

func NewHistoryStorageConfig(apiKey string) HistoryStorageConfig {
	return &historyStorageConfig{apiKey: apiKey}
}
