package cmd

import (
	"log/slog"
	"os"

	"github.com/spf13/cobra"
	"github.com/thyyl/chatr/internal/wire"
)

var matchCommand = &cobra.Command{
	Use:   "match",
	Short: "Match Server",
	Run: func(cmd *cobra.Command, args []string) {
		server, err := wire.InitializeMatchServer("match")
		if err != nil {
			slog.Error(err.Error())
			os.Exit(1)
		}
		server.Serve()
	},
}

func init() {
	rootCommand.AddCommand(matchCommand)
}
