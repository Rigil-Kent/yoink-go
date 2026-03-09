package cli

import (
	"fmt"
	"net/http"
	"os"

	"github.com/spf13/cobra"
)

var healthcheckCmd = &cobra.Command{
	Use:    "healthcheck",
	Short:  "Check if the web server is running (used by Docker HEALTHCHECK)",
	Args:   cobra.NoArgs,
	Hidden: true,
	Run: func(cmd *cobra.Command, args []string) {
		port, _ := cmd.Flags().GetString("port")
		resp, err := http.Get(fmt.Sprintf("http://localhost:%s/health", port))
		if err != nil || resp.StatusCode != http.StatusOK {
			os.Exit(1)
		}
	},
}

func init() {
	healthcheckCmd.Flags().StringP("port", "p", "8080", "Port the server is listening on")
	cli.AddCommand(healthcheckCmd)
}
