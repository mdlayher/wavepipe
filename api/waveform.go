package api

import (
	"database/sql"
	"fmt"
	"image/color"
	"image/png"
	"log"
	"net/http"
	"net/url"
	"strconv"

	"github.com/mdlayher/wavepipe/data"

	"github.com/gorilla/context"
	"github.com/gorilla/mux"
	"github.com/hashicorp/golang-lru"
	"github.com/mdlayher/waveform"
	"github.com/nfnt/resize"
	"github.com/unrolled/render"
)

// waveformLRU is a LRU cache which stores a fixed number of recently computed waveform image
// values, and evicts the least-recently-used entries when they are not used
var waveformLRU *lru.Cache

func init() {
	// Initialize fixed-capacity LRU cache
	var err error
	waveformLRU, err = lru.New(20)
	if err != nil {
		panic(err)
	}
}

// GetWaveform generates and returns a waveform image from wavepipe.  On success, this API will
// return a binary stream. On failure, it will return a JSON error.
func GetWaveform(w http.ResponseWriter, r *http.Request) {
	// Retrieve render
	ren := context.Get(r, CtxRender).(*render.Render)

	// Check API version
	if version, ok := mux.Vars(r)["version"]; ok {
		// Check if this API call is supported in the advertised version
		if !apiVersionSet.Has(version) {
			ren.JSON(w, 400, errRes(400, "unsupported API version: "+version))
			return
		}
	}

	// Check for an ID parameter
	pID, ok := mux.Vars(r)["id"]
	if !ok {
		ren.JSON(w, 400, errRes(400, "no integer song ID provided"))
		return
	}

	// Verify valid integer ID
	id, err := strconv.Atoi(pID)
	if err != nil {
		ren.JSON(w, 400, errRes(400, "invalid integer song ID"))
		return
	}

	// Check for a pre-existing set of waveform values
	var values []float64
	if tmpValues, ok := waveformLRU.Get(id); ok {
		// Use existing values
		values = tmpValues.([]float64)
	} else {
		// No existing waveform
		// Attempt to load the song with matching ID
		song := &data.Song{ID: id}
		if err := song.Load(); err != nil {
			// Check for invalid ID
			if err == sql.ErrNoRows {
				ren.JSON(w, 404, errRes(404, "song ID not found"))
				return
			}

			// All other errors
			log.Println(err)
			ren.JSON(w, 500, serverErr)
			return
		}

		// Open song's backing stream
		stream, err := song.Stream()
		if err != nil {
			log.Println(err)
			ren.JSON(w, 500, serverErr)
			return
		}

		// Compute waveform values from this song
		tmpValues, err := waveform.ComputeValues(stream, &waveform.ComputeOptions{
			Resolution: 4,
		})
		if err != nil {
			// If unknown format, return JSON error
			if err == waveform.ErrFormat {
				ren.JSON(w, 501, errRes(501, "unsupported audio format"))
				return
			}

			log.Println(err)
			ren.JSON(w, 500, serverErr)
			return
		}

		// Cache values, and use for generating current and future images
		waveformLRU.Add(id, tmpValues)
		values = tmpValues
	}

	// Check for optional color parameters
	var cR, cG, cB uint8

	// Background color
	var bgColor color.Color = color.White
	if bgColorStr := r.URL.Query().Get("bg"); bgColorStr != "" {
		// Convert %23 to #
		bgColorStr, err := url.QueryUnescape(bgColorStr)
		if err == nil {
			cR, cG, cB = hexToRGB(bgColorStr)
			bgColor = color.RGBA{cR, cG, cB, 255}
		}
	}

	// Foreground color
	var fgColor color.Color = color.Black
	if fgColorStr := r.URL.Query().Get("fg"); fgColorStr != "" {
		// Convert %23 to #
		fgColorStr, err := url.QueryUnescape(fgColorStr)
		if err == nil {
			cR, cG, cB = hexToRGB(fgColorStr)
			fgColor = color.RGBA{cR, cG, cB, 255}
		}
	}

	// Alternate color; follow foreground color by default
	var altColor color.Color = fgColor
	if altColorStr := r.URL.Query().Get("alt"); altColorStr != "" {
		// Convert %23 to #
		altColorStr, err := url.QueryUnescape(altColorStr)
		if err == nil {
			cR, cG, cB = hexToRGB(altColorStr)
			altColor = color.RGBA{cR, cG, cB, 255}
		}
	}

	// If requested, resize the image to the specified width
	var sizeX, sizeY int
	if strSize := r.URL.Query().Get("size"); strSize != "" {
		// Check for dimensions in two integers
		if _, err := fmt.Sscanf(strSize, "%dx%d", &sizeX, &sizeY); err != nil {
			ren.JSON(w, 400, errRes(400, "invalid x-separated integer pair for size"))
			return
		}
	}

	// Generate waveform image from computed values, with specified options
	img := waveform.DrawImage(values, &waveform.ImageOptions{
		ForegroundColor: fgColor,
		BackgroundColor: bgColor,
		AlternateColor:  altColor,

		ScaleX: 5,
		ScaleY: 4,

		Sharpness: 1,

		ScaleClipping: true,
	})

	// If a resize option was set, perform it now
	if sizeX > 0 {
		// Perform image resize
		img = resize.Resize(uint(sizeX), uint(sizeY), img, resize.NearestNeighbor)
	}

	// Encode as PNG to HTTP writer
	if err := png.Encode(w, img); err != nil {
		log.Println(err)
	}
}

// hexToRGB converts a hex string to a RGB triple.
// Credit: https://code.google.com/p/gorilla/source/browse/color/hex.go?r=ef489f63418265a7249b1d53bdc358b09a4a2ea0
func hexToRGB(h string) (uint8, uint8, uint8) {
	if len(h) > 0 && h[0] == '#' {
		h = h[1:]
	}
	if len(h) == 3 {
		h = h[:1] + h[:1] + h[1:2] + h[1:2] + h[2:] + h[2:]
	}
	if len(h) == 6 {
		if rgb, err := strconv.ParseUint(string(h), 16, 32); err == nil {
			return uint8(rgb >> 16), uint8((rgb >> 8) & 0xFF), uint8(rgb & 0xFF)
		}
	}
	return 0, 0, 0
}
