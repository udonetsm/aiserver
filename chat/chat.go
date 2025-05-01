package chat

import (
	"context"
	"fmt"

	"github.com/google/generative-ai-go/genai"
	"gitverse.ru/udonetsm/aiserver/historystorage"
	"gitverse.ru/udonetsm/aiserver/logger"
	"google.golang.org/api/iterator"
)

type chat struct {
	historyStorage historystorage.HistoryStorage
	*genai.ChatSession
	client   Client
	indxList []uint
	logger   logger.Logger
}

type Chat interface {
	SendMessage(ctx context.Context, message string, answer chan<- string) error
	SaveClient(client Client)
	Client() Client
	HistManager() HistoryManager
}

func (c *chat) SendMessage(ctx context.Context, message string, answer chan<- string) error {
	iter := c.ChatSession.SendMessageStream(ctx, genai.Text(message))
	for {
		select {
		case <-ctx.Done():
			close(answer)
			return fmt.Errorf("%w", ctx.Err())
		default:
			resp, err := iter.Next()
			if err == iterator.Done {
				close(answer)
				return nil
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

	}
}

func (c *chat) SaveClient(client Client) {
	c.client = client
}

func (c *chat) Client() Client {
	return c.client
}
