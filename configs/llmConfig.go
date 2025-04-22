package configs

import (
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

func (l *llmconfig) ApiKey() string {
	return l.apikey
}

func (l *llmconfig) ModelName() string {
	return l.modelname
}

func NewLLMConfig(logger logger.Logger, apikey, modelname string) (LLMConfig, error) {
	llmConfig := &llmconfig{logger: logger, apikey: apikey, modelname: modelname}
	return llmConfig, nil
}
