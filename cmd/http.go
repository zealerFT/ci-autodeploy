package cmd

import (
	"autodeploy/http"

	"github.com/rs/zerolog/log"

	"github.com/spf13/cobra"
)

var httpCmd = &cobra.Command{
	Use:   "http",
	Short: "start http server",
	PersistentPreRun: func(cmd *cobra.Command, args []string) {

	},
	Run: func(cmd *cobra.Command, args []string) {
		log.Info().Msg("autodeploy:boot api server")
		bootHTTPServer()
	},
}

func bootHTTPServer() {
	server := http.New(
		// SkipPaths
		http.ExportLogOption(),
		// whitelist
		http.WhitelistCheck(),
		// prometheus
		http.SetPrometheusMetrics(),
		// 加载route
		http.SetRouteOption(),
	)
	if err := server.Run(":8000"); err != nil {
		log.Panic().Msgf("HTTPServer run failed, err: %s", err)
	}
}
