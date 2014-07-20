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

// GetStatus returns the current server status, with an HTTP status and JSON
func GetStatus(res http.ResponseWriter, req *http.Request) {
	// Retrieve render
	r := context.Get(req, CtxRender).(*render.Render)

	// Output struct for songs request
	out := StatusResponse{}

	// Check API version
	if version, ok := mux.Vars(req)["version"]; ok {
		// Check if this API call is supported in the advertised version
		if !apiVersionSet.Has(version) {
			r.JSON(res, 400, errRes(400, "unsupported API version: "+version))
			return
		}
	}

	// Retrieve current server status
	status, err := common.ServerStatus()
	if err != nil {
		log.Println(err)
		r.JSON(res, 500, serverErr)
		return
	}

	// Copy into response
	out.Status = status

	// If requested, fetch additional metrics (not added by default due to full table scans in database)
	if metricTypes := req.URL.Query().Get("metrics"); metricTypes != "" {
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
					r.JSON(res, 500, serverErr)
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
	r.JSON(res, 200, out)
	return
}
