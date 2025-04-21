package chat

import (
	"context"
	"fmt"

	"github.com/google/generative-ai-go/genai"
	"gitverse.ru/udonetsm/aiserver/configs"
	"google.golang.org/api/option"
)

type client struct {
	config configs.LLMConfig
	*genai.Client
}

type Client interface {
	SendFile(ctx context.Context, path, name, mime string) (link string, err error)
	Generative() (Chat, error)
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

func (l *client) Generative() (Chat, error) {
	c := l.Client.GenerativeModel(l.config.ModelName())
	if c == nil {
		return nil, fmt.Errorf("nil chat session not allowed")
	}
	return &chat{ChatSession: c.StartChat()}, nil
}

func NewLLMClient(ctx context.Context, config configs.LLMConfig) (Client, error) {
	llmClient := &client{config: config}
	client, err := genai.NewClient(ctx, option.WithAPIKey(config.ApiKey()))
	if err != nil {
		return nil, fmt.Errorf("%w", err)
	}
	llmClient.Client = client
	return llmClient, nil
}
