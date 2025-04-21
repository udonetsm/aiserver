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
	Source() string
}

func (r *rootCmd) defaultSource() error {
	home := os.Getenv("HOME")
	if home == "" {
		return fmt.Errorf("empty $HOME variable")
	}
	r.envSource = filepath.Join(home, ".config.env")
	return nil
}

func (r *rootCmd) FetchCMD() error {
	r.Use = "aiserver"
	r.Short = "starts aiserver app"
	r.Run = func(cmd *cobra.Command, args []string) {}
	err := r.defaultSource()
	if err != nil {
		return fmt.Errorf("%w", err)
	}
	r.Command.Flags().StringVarP(&r.envSource, "source", "s", r.envSource, "set environment variable source as absolute path")
	r.Command.Execute()
	return nil
}

func (r *rootCmd) Source() string {
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
