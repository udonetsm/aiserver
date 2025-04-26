package sessions

import (
	"fmt"
	"sync"

	"gitverse.ru/udonetsm/aiserver/chat"
	"gitverse.ru/udonetsm/aiserver/logger"
)

type sessionStorage struct {
	logger  logger.Logger
	storage map[string]chat.Chat
	mu      sync.RWMutex
}

type SessionStorage interface {
	NewSession(key string, session chat.Chat) error
	SessionByKey(key string) (chat.Chat, error)
	DropSessionByKey(key string) error
}

func (s *sessionStorage) NewSession(key string, session chat.Chat) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	if key == "" {
		return fmt.Errorf("empty key not allowed")
	}
	_, ok := s.storage[key]
	if ok {
		return fmt.Errorf("session exists")
	}
	s.storage[key] = session
	return nil
}

func (s *sessionStorage) SessionByKey(key string) (chat.Chat, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	if key == "" {
		return nil, fmt.Errorf("empty key not allowed")
	}
	session, ok := s.storage[key]
	if !ok {
		return nil, fmt.Errorf("no session found by %v", key)
	}
	return session, nil
}

func (s *sessionStorage) DropSessionByKey(key string) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	if key == "" {
		return fmt.Errorf("empty key not allowed")
	}
	delete(s.storage, key)
	return nil
}

func NewSessionStorage(logger logger.Logger) SessionStorage {
	return &sessionStorage{logger: logger, storage: make(map[string]chat.Chat, 2)}
}
