package configs

import (
	"fmt"
	"os"
)

type grpcConfig struct {
	addr string
}

type GRPCConfig interface {
	GRPCAddr() string
}

func (g *grpcConfig) loadAddr() error {
	g.addr = os.Getenv("GRPCADDR")
	if g.addr == "" {
		return fmt.Errorf("empty grpc address not allowed")
	}
	return nil
}

func (g *grpcConfig) GRPCAddr() string {
	return g.addr
}

func NewGRPCConfig() (GRPCConfig, error) {
	grpcConfig := &grpcConfig{}
	err := grpcConfig.loadAddr()
	if err != nil {
		return nil, fmt.Errorf("%w", err)
	}
	return grpcConfig, nil
}
