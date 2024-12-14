package cmd

import (
	"log/slog"
	"os"

	"github.com/spf13/cobra"
	"github.com/thyyl/chatr/internal/wire"
)

var uploaderCommand = &cobra.Command{
	Use:   "uploader",
	Short: "Uploader Server",
	Run: func(cmd *cobra.Command, args []string) {
		server, err := wire.InitializeUploaderServer("uploader")
		if err != nil {
			slog.Error(err.Error())
			os.Exit(1)
		}
		server.Serve()
	},
}

func init() {
	rootCommand.AddCommand(uploaderCommand)
}
