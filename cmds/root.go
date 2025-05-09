package cmds

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/spf13/cobra"
)

type rootCmd struct {
	envSource string
	*cobra.Command
}

type RootCMD interface {
	FetchCMD() error
	EnvSource() string
}

func (r *rootCmd) homePath() (string, error) {
	home := os.Getenv("HOME")
	if home == "" {
		return "", fmt.Errorf("empty $HOME variable")
	}
	return home, nil
}

func (r *rootCmd) FetchCMD() error {
	r.Use = "aiserver"
	r.Short = "starts aiserver app"
	r.Run = func(cmd *cobra.Command, args []string) {}
	home, err := r.homePath()
	if err != nil {
		return fmt.Errorf("%w", err)
	}
	r.envSource = filepath.Join(home, ".config.env")
	r.Command.Flags().StringVarP(&r.envSource, "envsource", "e", r.envSource, "set where is your config.env file")
	r.Command.Execute()
	return nil
}

func (r *rootCmd) EnvSource() string {
	return r.envSource
}

func NewRootCMD() (RootCMD, error) {
	rootCmd := &rootCmd{Command: &cobra.Command{}}
	err := rootCmd.FetchCMD()
	if err != nil {
		return nil, err
	}
	return rootCmd, nil
}
