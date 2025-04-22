package infrastructure

import (
	"fmt"
	"net"

	"gitverse.ru/udonetsm/aiserver/configs"
	"gitverse.ru/udonetsm/aiserver/logger"
	"google.golang.org/grpc"
)

type server struct {
	logger     logger.Logger
	server     *grpc.Server
	listener   net.Listener
	grpcConfig configs.GRPCConfig
}

type Server interface {
	Server() *grpc.Server
	Listener() net.Listener
	Stop()
}

func (s *server) Server() *grpc.Server {
	return s.server
}

func (s *server) Listener() net.Listener {
	return s.listener
}

func (s *server) Stop() {
	s.server.GracefulStop()
}

func NewGRPCServer(logger logger.Logger, config configs.GRPCConfig) (Server, error) {
	grpcServer := &server{logger: logger, grpcConfig: config}
	grpcServer.server = grpc.NewServer()
	listener, err := net.Listen("tcp", grpcServer.grpcConfig.GRPCAddr())
	if err != nil {
		return nil, fmt.Errorf("%w", err)
	}
	grpcServer.listener = listener
	grpcServer.logger.Info("server configured on ", grpcServer.grpcConfig.GRPCAddr())
	return grpcServer, nil
}
