package chat

import (
	"context"
	"fmt"

	"github.com/google/generative-ai-go/genai"
	"gitverse.ru/udonetsm/aiserver/logger"
	"google.golang.org/api/iterator"
)

type chat struct {
	*genai.ChatSession
	client   Client
	indxList []uint
	logger   logger.Logger
}

type Chat interface {
	SendMessage(ctx context.Context, message string, answer chan<- string) error
	SetClient(client Client)
	Client() Client
	HistManager() HistoryManager
}

type HistoryManager interface {
	AddMessageToHistory(ctx context.Context, message, role, ctype string) (uint, error)
	SaveFileIndex(indx uint) error
	DropFileInfoMessageByIndex(indx uint) error
	HistoryFileIndexes() ([]uint, error)
	ClearHistory(ctx context.Context) error
}

func (c *chat) HistManager() HistoryManager {
	return c
}

func (c *chat) DropFileInfoMessageByIndex(indx uint) error {
	if c.ChatSession.History == nil {
		return fmt.Errorf("nil chat session history not allowed")
	}

	if uint(len(c.ChatSession.History)) <= indx {
		return fmt.Errorf("elemet not exixts")
	}

	c.ChatSession.History[indx] =
		&genai.Content{
			Parts: []genai.Part{
				genai.Text("d"),
			},
			Role: "user"}
	c.logger.Infof("removed message by %v indx", indx)
	return nil
}

func (c *chat) HistoryFileIndexes() ([]uint, error) {
	var err error
	if len(c.ChatSession.History) < 1 {
		err = fmt.Errorf("history cleared. Skip clearing indexes")
		c.indxList = make([]uint, 0)
	}
	return c.indxList, err
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

func (c *chat) AddMessageToHistory(ctx context.Context, message, role, ctype string) (uint, error) {
	if c.ChatSession == nil {
		return 0, fmt.Errorf("nil chat session not allowed")
	}
	if c.ChatSession.History == nil {
		return 0, fmt.Errorf("nil history not allowed")
	}
	var content *genai.Content
	if ctype != "" {
		content = &genai.Content{
			Parts: []genai.Part{genai.FileData{
				MIMEType: ctype,
				URI:      message,
			}},
			Role: role,
		}
	} else {
		content = &genai.Content{
			Parts: []genai.Part{genai.Text(message)},
			Role:  role,
		}
	}
	c.ChatSession.History = append(c.ChatSession.History, content)
	return uint(len(c.ChatSession.History) - 1), nil
}

func (c *chat) SaveFileIndex(indx uint) error {
	if c.indxList == nil {
		return fmt.Errorf("nil index list not allowed")
	}
	c.indxList = append(c.indxList, indx)
	return nil
}

func (c *chat) ClearHistory(ctx context.Context) error {
	if c.ChatSession == nil {
		return fmt.Errorf("nil chat session not allowed")
	}
	if c.ChatSession.History == nil || len(c.ChatSession.History) < 1 {
		return fmt.Errorf("nil or empty chat history not allowed")
	}
	c.ChatSession.History = make([]*genai.Content, 0)
	return nil
}

func (c *chat) SetClient(client Client) {
	c.client = client
}

func (c *chat) Client() Client {
	return c.client
}
