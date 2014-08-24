package api

import (
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/mdlayher/wavepipe/common"

	"github.com/gorilla/context"
	"github.com/gorilla/mux"
	"github.com/mdlayher/goset"
	"github.com/unrolled/render"
)

// databaseMetricsCache is used to cache database metrics until a database update occurs
var databaseMetricsCache *common.DatabaseMetrics

// cacheTime is used to determine when metrics were last cached
var cacheTime int64

// StatusResponse represents the JSON response for /api/status
type StatusResponse struct {
	Error   *Error          `json:"error"`
	Status  *common.Status  `json:"status"`
	Metrics *common.Metrics `json:"metrics"`
}

// GetStatus returns the current server status, and optionally, server metrics, with
// an HTTP status and JSON.
func GetStatus(w http.ResponseWriter, r *http.Request) {
	// Retrieve render
	ren := context.Get(r, CtxRender).(*render.Render)

	// Output struct for songs request
	out := StatusResponse{}

	// Check API version
	if version, ok := mux.Vars(r)["version"]; ok {
		// Check if this API call is supported in the advertised version
		if !apiVersionSet.Has(version) {
			ren.JSON(w, 400, errRes(400, "unsupported API version: "+version))
			return
		}
	}

	// Retrieve current server status
	out.Status = common.ServerStatus()

	// If requested, fetch additional metrics (not added by default due to full table scans in database)
	if metricTypes := r.URL.Query().Get("metrics"); metricTypes != "" {
		// Begin building metrics
		metrics := &common.Metrics{}

		// Constants to check for various metric types
		const (
			mAll      = "all"
			mDatabase = "database"
			mNetwork  = "network"
		)

		// Set of valid metric types
		validSet := set.New(mAll, mDatabase, mNetwork)

		// Check for comma-separated list of metric types
		metricSet := set.New()
		for _, m := range strings.Split(metricTypes, ",") {
			// Add valid types to set
			if validSet.Has(m) {
				metricSet.Add(m)
			}
		}

		// If requested, get metrics about the database
		if metricSet.Has(mAll) || metricSet.Has(mDatabase) {
			// Check for cached metrics, and make sure they are up to date
			if databaseMetricsCache != nil && common.ScanTime() <= cacheTime {
				metrics.Database = databaseMetricsCache
			} else {
				// Fetch new metrics
				dbMetrics, err := common.GetDatabaseMetrics()
				if err != nil {
					log.Println(err)
					ren.JSON(w, 500, serverErr)
					return
				}
				metrics.Database = dbMetrics

				// Cache metrics and update time
				databaseMetricsCache = dbMetrics
				cacheTime = time.Now().Unix()
			}
		}

		// If requested, get metrics about the network
		if metricSet.Has(mAll) || metricSet.Has(mNetwork) {
			metrics.Network = &common.NetworkMetrics{
				RXBytes: common.RXBytes(),
				TXBytes: common.TXBytes(),
			}
		}

		// Return metrics
		out.Metrics = metrics
	}

	// HTTP 200 OK with JSON
	out.Error = nil
	ren.JSON(w, 200, out)
	return
}
