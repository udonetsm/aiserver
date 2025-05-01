package contentreader

import (
	"context"
	"fmt"
	"net/http"
	"os"

	"gitverse.ru/udonetsm/aiserver/configs"
	"gitverse.ru/udonetsm/aiserver/logger"
)

type contentReader struct {
	logger                    logger.Logger
	contentReaderReaderConfig configs.ContentReaderConfig
}

type ContentReader interface {
	Configure(source string) error
	ReadContent(ctx context.Context) ([]byte, error)
	DetectContentType(ctx context.Context, content []byte) (string, error)
}

func (fr *contentReader) Configure(source string) error {
	return fr.Configure(source)
}

func (fr *contentReader) ReadContent(ctx context.Context) ([]byte, error) {
	read, err := os.ReadFile(fr.contentReaderReaderConfig.ContentSource())
	if err != nil {
		return nil, fmt.Errorf("clien.readfile error: %w", err)
	}
	if len(read) < 1 {
		return nil, fmt.Errorf("empty file")
	}
	return read, nil
}

func (fr *contentReader) DetectContentType(ctx context.Context, content []byte) (string, error) {
	ct := http.DetectContentType(content)
	if ct == "" {
		return "", fmt.Errorf("content type not detected")
	}
	return ct, nil
}

func NewContentReader(logger logger.Logger, contentReaderConfig configs.ContentReaderConfig) ContentReader {
	return &contentReader{logger: logger, contentReaderReaderConfig: contentReaderConfig}
}
