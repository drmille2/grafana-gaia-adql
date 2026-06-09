package main

import (
	"os"

	"github.com/grafana/grafana-plugin-sdk-go/backend/datasource"
	"github.com/grafana/grafana-plugin-sdk-go/backend/log"

	"github.com/davidmillerpf9/grafana-gaia-adql/pkg/plugin"
)

func main() {
	// Start listening to requests sent from the Grafana backend. The
	// grafana-plugin-sdk-go SDK takes care of wiring the plugin to the
	// Grafana gRPC server and provides a convenient ServeOpts struct to
	// configure the plugin.
	if err := datasource.Manage("davidmillerpf9-gaia-adql-datasource", plugin.NewGaiaADQLDatasource, datasource.ManageOpts{}); err != nil {
		log.DefaultLogger.Error("error managing plugin", "error", err)
		os.Exit(1)
	}
}
