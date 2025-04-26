package chat

import (
	"context"
	"fmt"

	"github.com/google/generative-ai-go/genai"
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
	SendFile(ctx context.Context, path, name, mime string) (link string, err error)
	LisFiles(ctx context.Context) ([]string, error)
	DeleteFileByFilename(ctx context.Context, filename string) error
	Generative() Model
}

func (c *client) SendFile(ctx context.Context, path, name, mime string) (link string, err error) {
	opts := &genai.UploadFileOptions{
		DisplayName: name,
		MIMEType:    mime,
	}
	genaiFile, err := c.Client.UploadFileFromPath(ctx, path, opts)
	if err != nil {
		return "", fmt.Errorf("%w", err)
	}
	return genaiFile.URI, nil
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
