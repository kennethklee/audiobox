package cmd

import (
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/pocketbase/pocketbase/core"
	"github.com/spf13/cobra"
)

func NewHealthCheckCmd(app core.App) *cobra.Command {
	return &cobra.Command{
		Use:   "healthcheck",
		Short: "Check the health",
		Run: func(cmd *cobra.Command, args []string) {
			exitCode := 0

			fmt.Println("Health Check")
			fmt.Println("  Database")
			if err := app.DB().DB().Ping(); err != nil {
				log.Println("    Error:", err)
				exitCode = 1
			} else {
				fmt.Println("    OK")
			}

			fmt.Println("  Server")
			resp, err := http.Get("http://localhost:8090/api/health")
			if err != nil {
				fmt.Println("    Error:", err)
				exitCode = 1
			} else if resp.StatusCode != 200 {
				err := fmt.Errorf("status code %d", resp.StatusCode)
				fmt.Println("    Error:", err)
				exitCode = 1
			} else {
				fmt.Println("    OK")
			}

			os.Exit(exitCode)
		},
	}
}
