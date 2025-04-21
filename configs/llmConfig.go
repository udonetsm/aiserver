package configs

import (
	"fmt"
	"os"

	"gitverse.ru/udonetsm/aiserver/logger"
)

type llmconfig struct {
	apikey, modelname string
	logger            logger.Logger
}

type LLMConfig interface {
	ApiKey() string
	ModelName() string
}

func (l *llmconfig) setApikeyEnv() error {
	l.apikey = os.Getenv("APIKEY")
	if l.apikey == "" {
		return fmt.Errorf("empty api key not allowed")
	}
	l.logger.Info("success set api key", l.apikey)
	return nil
}
func (l *llmconfig) setModelNameEnv() error {
	l.modelname = os.Getenv("MODELNAME")
	if l.modelname == "" {
		return fmt.Errorf("empty model name not allowed")
	}
	l.logger.Info("success set model name", l.modelname)
	return nil
}

func (l *llmconfig) ApiKey() string {
	return l.apikey
}

func (l *llmconfig) ModelName() string {
	return l.modelname
}

func NewLLMConfig(logger logger.Logger) (LLMConfig, error) {
	llmConfig := &llmconfig{logger: logger}
	err := llmConfig.setApikeyEnv()
	if err != nil {
		return nil, fmt.Errorf("%w", err)
	}
	err = llmConfig.setModelNameEnv()
	if err != nil {
		return nil, fmt.Errorf("%w", err)
	}
	return llmConfig, nil
}
