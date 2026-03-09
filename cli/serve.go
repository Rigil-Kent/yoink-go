package cli

import (
	"fmt"
	"log"
	"os"
	"path/filepath"

	"github.com/spf13/cobra"
	"yoink/web"
)

var serveCmd = &cobra.Command{
	Use:   "serve",
	Short: "Start the Yoink web UI",
	Args:  cobra.NoArgs,
	Run: func(cmd *cobra.Command, args []string) {
		library, ok := os.LookupEnv("YOINK_LIBRARY")
		if !ok {
			userHome, _ := os.UserHomeDir()
			library = filepath.Join(userHome, ".yoink")
		}

		port, _ := cmd.Flags().GetString("port")
		addr := fmt.Sprintf(":%s", port)

		if err := web.Listen(addr, library); err != nil {
			log.Fatal(err)
		}
	},
}

func init() {
	serveCmd.Flags().StringP("port", "p", "8080", "Port to listen on")
	cli.AddCommand(serveCmd)
}
