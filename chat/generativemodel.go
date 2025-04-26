package chat

import (
	"github.com/google/generative-ai-go/genai"
	"gitverse.ru/udonetsm/aiserver/logger"
)

type generativeModel struct {
	*genai.GenerativeModel
}

type Model interface {
	Start(logger.Logger) Chat
}

func (g *generativeModel) Start(logger logger.Logger) Chat {
	c := &chat{ChatSession: g.GenerativeModel.StartChat()}
	c.ChatSession.History = make([]*genai.Content, 0)
	c.indxList = make([]uint, 0)
	c.logger = logger
	return c
}
