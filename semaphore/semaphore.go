package semaphore

import "gitverse.ru/udonetsm/aiserver/configs"

type semaphore struct {
	taskAmount chan struct{}
}

type Semaphore interface {
	Acquire()
	Release()
}

func (s *semaphore) Acquire() {
	s.taskAmount <- struct{}{}
}

func (s *semaphore) Release() {
	<-s.taskAmount
}

func NewSemaphore(semaphoreConfig configs.SemaphoreConfig) Semaphore {
	return &semaphore{taskAmount: make(chan struct{}, semaphoreConfig.TaskAmount())}
}
