package configs

type llmconfig struct {
	apikey, modelname string
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

func NewLLMConfig(apikey, modelname string) (LLMConfig, error) {
	llmConfig := &llmconfig{apikey: apikey, modelname: modelname}
	return llmConfig, nil
}
