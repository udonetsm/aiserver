package bootstrap

import (
	ai_ "gitverse.ru/udonetsm/aiserver/aipack"
	"gitverse.ru/udonetsm/aiserver/cmds"
	"gitverse.ru/udonetsm/aiserver/configs"
	"gitverse.ru/udonetsm/aiserver/envloader"
	"gitverse.ru/udonetsm/aiserver/handlers"
	"gitverse.ru/udonetsm/aiserver/historystorage"
	"gitverse.ru/udonetsm/aiserver/infrastructure"
	"gitverse.ru/udonetsm/aiserver/logger"
	"gitverse.ru/udonetsm/aiserver/semaphore"
	"gitverse.ru/udonetsm/aiserver/sessions"
)

type bootstrap struct {
	logger            logger.Logger
	semConfig         configs.SemaphoreConfig
	semaphore         semaphore.Semaphore
	rootCMD           cmds.RootCMD
	envLoader         envloader.EnvLoader
	sessionStorage    sessions.SessionStorage
	handlers          handlers.Handlers
	grpcConfig        configs.GRPCConfig
	server            infrastructure.Server
	histStorageConfig configs.HistoryStorageConfig
	historyStorage    historystorage.HistoryStorage
}

type Bootstrap interface {
	Load()
}

func (b *bootstrap) Load() {
	var err error

	b.logger = logger.NewLogger()

	b.rootCMD, err = cmds.NewRootCMD()
	if err != nil {
		b.logger.Fatal(err)
	}

	b.envLoader = envloader.NewEnvLoader(b.rootCMD.EnvSource(), b.logger)
	err = b.envLoader.LoadEnvs()
	if err != nil {
		b.logger.Fatal(err)
	}

	b.semConfig, err = configs.NewSemaphoreConfig()
	if err != nil {
		b.logger.Info(err)
	}

	b.sessionStorage = sessions.NewSessionStorage(b.logger)

	b.handlers = handlers.NewHandlers(b.logger, b.sessionStorage, b.semConfig, b.rootCMD.HistorySource())

	b.grpcConfig, err = configs.NewGRPCConfig()
	if err != nil {
		b.logger.Fatal(err)
	}

	b.server, err = infrastructure.NewGRPCServer(b.logger, b.grpcConfig)
	if err != nil {
		b.logger.Fatal(err)
	}
	ai_.RegisterTransmitServiceServer(b.server.Server(), b.handlers)

	b.server.Server().Serve(b.server.Listener())

}

func NewBootstrap() Bootstrap {
	return &bootstrap{}
}
