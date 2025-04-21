package bootstrap

import (
	"context"

	"gitverse.ru/udonetsm/aiserver/chat"
	"gitverse.ru/udonetsm/aiserver/cmds"
	"gitverse.ru/udonetsm/aiserver/configs"
	envloader "gitverse.ru/udonetsm/aiserver/envLoader"
	"gitverse.ru/udonetsm/aiserver/logger"
)

type bootstrap struct {
	logger    logger.Logger
	rootCMD   cmds.RootCMD
	envLoader envloader.EnvLoader
	llmConfig configs.LLMConfig
	llmClient chat.Client
	chat      chat.Chat
}

type Bootstrap interface {
	Load()
}

func (b *bootstrap) Load() {
	ctx := context.Background()
	var err error

	b.logger = logger.NewLogger()

	b.rootCMD, err = cmds.NewRootCMD()
	if err != nil {
		b.logger.Fatal(err)
	}
	b.envLoader = envloader.NewEnvLoader(b.rootCMD.Source(), b.logger)
	err = b.envLoader.LoadEnvs()
	if err != nil {
		b.logger.Fatal(err)
	}

	b.llmConfig, err = configs.NewLLMConfig(b.logger)
	if err != nil {
		b.logger.Fatal(err)
	}

	b.llmClient, err = chat.NewLLMClient(ctx, b.llmConfig)
	if err != nil {
		b.logger.Fatal(err)
	}

	b.chat, err = chat.NewChatSession(b.llmClient)
	if err != nil {
		b.logger.Fatal(err)
	}
}

func NewBootstrap() Bootstrap {
	return &bootstrap{}
}
