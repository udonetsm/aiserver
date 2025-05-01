package history

import (
	"context"
	"fmt"
)

type message struct {
	message, role string
}

type history struct {
	messsages []message
}

type History interface {
	BatchMessage(ctx context.Context, mess, role string, indx uint) error
	Remove() error
}

func (h *history) BatchMessage(ctx context.Context, mess, role string, indx uint) error {
	if h.messsages == nil {
		return fmt.Errorf("nil history not allowed")
	}
	h.messsages[indx] = message{message: mess, role: role}
	return nil
}

func (h *history) Remove() error {
	h.messsages = nil
	return nil
}

func NewHistory(historyLen uint) History {
	return &history{messsages: make([]message, historyLen)}
}
