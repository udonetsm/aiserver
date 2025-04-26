package chat

import "github.com/google/generative-ai-go/genai"

type generativeModel struct {
	*genai.GenerativeModel
}

type Model interface {
	Start() Chat
}

func (g *generativeModel) Start() Chat {
	c := &chat{ChatSession: g.GenerativeModel.StartChat()}
	c.ChatSession.History = make([]*genai.Content, 0)
	return c
}

func NewGenerativeModel() Model {
	return &generativeModel{}
}
