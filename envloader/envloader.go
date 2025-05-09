package envloader

import (
	"fmt"
	"log"
	"os"

	"github.com/joho/godotenv"
)

type envLoader struct {
	source string
}

type EnvLoader interface {
	LoadEnvs() error
}

func (e *envLoader) LoadEnvs() error {
	err := godotenv.Load(e.source)
	if err != nil {
		return fmt.Errorf("%w<---", err)
	}
	return nil
}

func NewEnvLoader(source string) EnvLoader {
	if _, err := os.Stat(source); err != nil {
		log.Fatal(err)
	}
	return &envLoader{source: source}
}
