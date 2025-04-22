package configs

import (
	"fmt"
	"os"
	"strconv"
)

type semaphoreConfig struct {
	taskAmount int
}

type SemaphoreConfig interface {
	TaskAmount() int
}

func (s *semaphoreConfig) setEnv() error {
	taskAmountText := os.Getenv("TASK_LIMIT")
	s.taskAmount, _ = strconv.Atoi(taskAmountText)
	if s.taskAmount == 0 {
		s.taskAmount = 5
		return fmt.Errorf("task limit not set; your TASK_LIMIT variable = %s ; default: 5", taskAmountText)
	}
	return nil
}

func (s *semaphoreConfig) TaskAmount() int {
	return s.taskAmount
}

func NewSemaphoreConfig() (SemaphoreConfig, error) {
	sc := &semaphoreConfig{}
	err := sc.setEnv()
	return sc, fmt.Errorf("%w", err)
}
