package chat

import (
	"context"
	"fmt"

	"github.com/google/generative-ai-go/genai"
	"google.golang.org/api/iterator"
)

type chat struct {
	llmClient Client
	*genai.ChatSession
}

type Chat interface {
	SendMessage(ctx context.Context, message string, answer chan<- string) error
}

func (c *chat) SendMessage(ctx context.Context, message string, answer chan<- string) error {
	iter := c.ChatSession.SendMessageStream(ctx, genai.Text(message))
	for {
		resp, err := iter.Next()
		if err == iterator.Done {
			close(answer)
			break
		}
		if err != nil {
			return fmt.Errorf("%w", err)
		}
		for _, cand := range resp.Candidates {
			if cand.Content != nil {
				for _, part := range cand.Content.Parts {
					answer <- fmt.Sprintf("%v", part)
				}
			}
		}
	}
	return nil
}

func NewChatSession(llmClient Client) (Chat, error) {
	c, err := llmClient.Generative()
	if err != nil {
		return nil, fmt.Errorf("%w", err)
	}
	return c, nil
}
