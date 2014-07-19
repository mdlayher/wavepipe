package api

import (
	"log"
	"net/http"
	"strings"

	"github.com/mdlayher/wavepipe/common"

	"github.com/gorilla/context"
	"github.com/gorilla/mux"
	"github.com/mdlayher/goset"
	"github.com/unrolled/render"
)

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
		)

		// Set of valid metric types
		validSet := set.New(mAll, mDatabase)

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
			dbMetrics, err := common.GetDatabaseMetrics()
			if err != nil {
				log.Println(err)
				r.JSON(res, 500, serverErr)
				return
			}
			metrics.Database = dbMetrics
		}

		// Return metrics
		out.Metrics = metrics
	}

	// HTTP 200 OK with JSON
	out.Error = nil
	r.JSON(res, 200, out)
	return
}
