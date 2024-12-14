package cmd

import (
	"log/slog"
	"os"

	"github.com/spf13/cobra"
	"github.com/thyyl/chatr/internal/wire"
)

var chatCommand = &cobra.Command{
	Use:   "chat",
	Short: "Chat Server",
	Run: func(cmd *cobra.Command, args []string) {
		server, err := wire.InitializeChatServer("chat")
		if err != nil {
			slog.Error(err.Error())
			os.Exit(1)
		}
		server.Serve()
	},
}

func init() {
	rootCommand.AddCommand(chatCommand)
}
