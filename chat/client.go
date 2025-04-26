package chat

import (
	"bytes"
	"context"
	"fmt"
	"net/http"
	"os"
	"strings"

	"github.com/google/generative-ai-go/genai"
	"github.com/google/uuid"
	"gitverse.ru/udonetsm/aiserver/configs"
	"gitverse.ru/udonetsm/aiserver/logger"
	"google.golang.org/api/iterator"
	"google.golang.org/api/option"
)

type client struct {
	config          configs.LLMConfig
	logger          logger.Logger
	semaphoreConfig configs.SemaphoreConfig
	*genai.Client
}

type Client interface {
	SendFile(ctx context.Context, path string) (link, ctype string, err error)
	LisFiles(ctx context.Context) ([]string, error)
	DeleteFileByFilename(ctx context.Context, filename string) error
	Generative() Model
}

func (c *client) readFile(adbsPath string) ([]byte, error) {
	read, err := os.ReadFile(adbsPath)
	if err != nil {
		return nil, fmt.Errorf("clien.readfile error: %w", err)
	}
	if len(read) < 1 {
		return nil, fmt.Errorf("empty file")
	}
	return read, nil
}

func (c *client) detectContentType(content []byte) (string, error) {
	ct := http.DetectContentType(content)
	if ct == "" {
		return "", fmt.Errorf("content type not detected")
	}
	return ct, nil
}

func (c *client) contentTypeSupported(contentType string) bool {
	return !strings.Contains(contentType, "octet") && !strings.Contains(contentType, "zip")
}

func (c *client) SendFile(ctx context.Context, path string) (link, ctype string, err error) {
	read, err := c.readFile(path)
	if err != nil {
		return "", "", fmt.Errorf("%w", err)
	}
	contentType, err := c.detectContentType(read)
	if err != nil {
		return "", "", fmt.Errorf("%w", err)
	}
	if !c.contentTypeSupported(contentType) {
		return "", "", fmt.Errorf("content type not supported")
	}
	name := uuid.NewString()
	opts := &genai.UploadFileOptions{
		DisplayName: name,
		MIMEType:    contentType,
	}
	genaiFile, err := c.Client.UploadFile(ctx, name, bytes.NewBuffer(read), opts)
	if err != nil {
		return "", "", fmt.Errorf("%w", err)
	}
	return genaiFile.URI, contentType, nil
}

func (c *client) LisFiles(ctx context.Context) ([]string, error) {
	fileInfos := make([]string, 0)
	iter := c.Client.ListFiles(ctx)
	for {
		file, err := iter.Next()
		if err == iterator.Done {
			if len(fileInfos) < 1 {
				return nil, fmt.Errorf("no files found for associated user")
			}
			return fileInfos, nil
		}
		if err != nil {
			return fileInfos, fmt.Errorf("%w", err)
		}
		fileInfos = append(fileInfos, file.Name)
	}
}

func (c *client) DeleteFileByFilename(ctx context.Context, filename string) error {

	c.logger.Infof("removing [%s]", filename)

	err := c.Client.DeleteFile(ctx, filename)
	if err != nil {
		return fmt.Errorf("%w", err)
	}
	c.logger.Infof("removed [%s]", filename)
	return nil
}

func (c *client) Generative() Model {
	gm := c.Client.GenerativeModel(c.config.ModelName())
	return &generativeModel{GenerativeModel: gm}
}

func NewClient(ctx context.Context, llmconfig configs.LLMConfig, logger logger.Logger, semaphoreConfig configs.SemaphoreConfig) (Client, error) {
	c, err := genai.NewClient(ctx, option.WithAPIKey(llmconfig.ApiKey()))
	if err != nil {
		return nil, fmt.Errorf("%s", err)
	}
	return &client{Client: c, config: llmconfig, logger: logger, semaphoreConfig: semaphoreConfig}, nil
}
