package chat

import (
	"context"
	"fmt"

	"github.com/google/generative-ai-go/genai"
	"gitverse.ru/udonetsm/aiserver/history"
	"gitverse.ru/udonetsm/aiserver/historystorage"
)

type HistoryManager interface {
	AddMessageToHistory(ctx context.Context, message, role, ctype string) (uint, error)
	SaveFileIndex(indx uint) error
	DropFileInfoMessageByIndex(indx uint) error
	HistoryFileIndexes() ([]uint, error)
	ClearHistory(ctx context.Context) error
	SaveHistory(ctx context.Context, historyStorage historystorage.HistoryStorage) error
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

func (c *chat) SaveHistory(ctx context.Context, historyStorage historystorage.HistoryStorage) error {
	history := history.NewHistory(uint(len(c.ChatSession.History)))

	for indx, message := range c.ChatSession.History {
		err := history.BatchMessage(ctx, fmt.Sprintf("%v", message.Parts), message.Role, uint(indx))
		if err != nil {
			return fmt.Errorf("chat.SaveHistory error: %w", err)
		}
	}
	err := historyStorage.Save(ctx, history)
	if err != nil {
		return fmt.Errorf("chat.SaveHistory error: %w", err)
	}
	return nil
}
