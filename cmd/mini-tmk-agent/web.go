package main

import (
	"fmt"
	"net/http"

	"github.com/lucyliuu/mini-tmk-agent/internal/web"
	"github.com/spf13/cobra"
)

func newWebCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "web",
		Short: "Start web UI server",
		RunE: func(cmd *cobra.Command, args []string) error {
			port, _ := cmd.Flags().GetInt("port")

			cfg := loadConfig()

			srv := web.NewServer(web.ServerConfig{
				APIKey:           cfg.APIKey,
				BaseURL:          cfg.BaseURL,
				DeepgramAPIKey:   cfg.DeepgramAPIKey,
				Port:             port,
				PublicBaseURL:    cfg.PublicBaseURL,
				LiveKitURL:       cfg.LiveKitURL,
				LiveKitAPIKey:    cfg.LiveKitAPIKey,
				LiveKitAPISecret: cfg.LiveKitAPISecret,
			})

			addr := fmt.Sprintf(":%d", port)
			fmt.Printf("Web UI running at http://localhost%s\n", addr)
			return http.ListenAndServe(addr, srv.Handler())
		},
	}

	cmd.Flags().Int("port", 8080, "HTTP port")
	return cmd
}
