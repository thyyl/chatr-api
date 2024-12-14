package cmd

import (
	"log/slog"
	"os"

	"github.com/spf13/cobra"
	"github.com/thyyl/chatr/internal/wire"
)

var forwarderCommand = &cobra.Command{
	Use:   "forwarder",
	Short: "Forwarder Server",
	Run: func(cmd *cobra.Command, args []string) {
		server, err := wire.InitializeForwarderServer("forwarder")
		if err != nil {
			slog.Error(err.Error())
			os.Exit(1)
		}
		server.Serve()
	},
}

func init() {
	rootCommand.AddCommand(forwarderCommand)
}