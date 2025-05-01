package chat

import (
	"context"
	"fmt"

	"github.com/google/generative-ai-go/genai"
	"gitverse.ru/udonetsm/aiserver/configs"
	"gitverse.ru/udonetsm/aiserver/contentreader"
	"gitverse.ru/udonetsm/aiserver/logger"
	"google.golang.org/api/option"
)

type client struct {
	config          configs.LLMConfig
	logger          logger.Logger
	semaphoreConfig configs.SemaphoreConfig
	contentReader   contentreader.ContentReader
	*genai.Client
}

type Client interface {
	Generative() Model
	FileManager() FileManager
}

func (c *client) FileManager() FileManager {
	return c
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
	return &client{Client: c, config: llmconfig,
		logger:          logger,
		semaphoreConfig: semaphoreConfig,
	}, nil
}
