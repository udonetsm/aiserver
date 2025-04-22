package envloader

import (
	"fmt"

	"github.com/joho/godotenv"
	"gitverse.ru/udonetsm/aiserver/logger"
)

type envLoader struct {
	source string
	logger logger.Logger
}

type EnvLoader interface {
	LoadEnvs() error
}

func (e *envLoader) LoadEnvs() error {
	err := godotenv.Load(e.source)
	if err != nil {
		return fmt.Errorf("%w", err)
	}
	return nil
}

func NewEnvLoader(source string, logger logger.Logger) EnvLoader {
	return &envLoader{source: source, logger: logger}
}
