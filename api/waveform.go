package api

import (
	"bytes"
	"database/sql"
	"fmt"
	"image/color"
	"image/png"
	"io"
	"log"
	"net/http"
	"net/url"
	"strconv"

	"github.com/mdlayher/wavepipe/data"

	"github.com/gorilla/context"
	"github.com/gorilla/mux"
	"github.com/mdlayher/waveform"
	"github.com/nfnt/resize"
	"github.com/unrolled/render"
)

const (
	// cacheThreshold is the number of waveform images which will be retained in-memory
	// after generation
	cacheThreshold = 20
)

// waveformCache stores encoded waveform images in-memory, for re-use
// through multiple HTTP calls
var waveformCache = map[string][]byte{}

// waveformList tracks insertion order for cached waveforms, and enables the removal
// of the oldest waveform once a threshold is reached
var waveformList = []string{}

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

	// Check for optional color parameters
	// Background color
	var bgColor color.Color = color.White
	if bgColorStr := r.URL.Query().Get("bg"); bgColorStr != "" {
		// Convert %23 to #
		bgColorStr, err := url.QueryUnescape(bgColorStr)
		if err == nil {
			cR, cG, cB := hexToRGB(bgColorStr)
			bgColor = color.RGBA{cR, cG, cB, 255}
		}
	}

	// Foreground color
	var fgColor color.Color = color.Black
	if fgColorStr := r.URL.Query().Get("fg"); fgColorStr != "" {
		// Convert %23 to #
		fgColorStr, err := url.QueryUnescape(fgColorStr)
		if err == nil {
			cR, cG, cB := hexToRGB(fgColorStr)
			fgColor = color.RGBA{cR, cG, cB, 255}
		}
	}

	// Alternate color; follow foreground color by default
	var altColor color.Color = fgColor
	if altColorStr := r.URL.Query().Get("alt"); altColorStr != "" {
		// Convert %23 to #
		altColorStr, err := url.QueryUnescape(altColorStr)
		if err == nil {
			cR, cG, cB := hexToRGB(altColorStr)
			altColor = color.RGBA{cR, cG, cB, 255}
		}
	}

	// Set up options struct for waveform
	options := &waveform.Options{
		ForegroundColor: fgColor,
		BackgroundColor: bgColor,
		AlternateColor:  altColor,

		Resolution: 4,

		ScaleX: 5,
		ScaleY: 4,

		Sharpness: 1,

		ScaleClipping: true,
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

	// Generate waveform cache key using ID, size, and options
	cacheKey := waveformCacheKey(id, sizeX, sizeY, options)

	// Check for a cached waveform
	if _, ok := waveformCache[cacheKey]; ok {
		// Send cached data to HTTP writer
		if _, err := io.Copy(w, bytes.NewReader(waveformCache[cacheKey])); err != nil {
			log.Println(err)
		}

		return
	}

	// Open song's backing stream
	stream, err := song.Stream()
	if err != nil {
		log.Println(err)
		ren.JSON(w, 500, serverErr)
		return
	}

	// Generate a waveform from this song
	img, err := waveform.New(stream, options)
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

	// If a resize option was set, perform it now
	if sizeX > 0 {
		// Perform image resize
		img = resize.Resize(uint(sizeX), uint(sizeY), img, resize.NearestNeighbor)
	}

	// Encode as PNG into buffer
	buf := bytes.NewBuffer(nil)
	if err := png.Encode(buf, img); err != nil {
		log.Println(err)
	}

	// Store cached image, append to cache list
	waveformCache[cacheKey] = buf.Bytes()
	waveformList = append(waveformList, cacheKey)

	// If threshold reached, remove oldest waveform from cache
	if len(waveformList) > cacheThreshold {
		oldest := waveformList[0]
		waveformList = waveformList[1:]
		delete(waveformCache, oldest)
	}

	// Send over HTTP
	if _, err := io.Copy(w, buf); err != nil {
		log.Println(err)
	}
}

// waveformCacheKey generates a cache key using waveform parameters, so that
// the waveform can be uniquely identified when cached
func waveformCacheKey(id int, sizeX int, sizeY int, options *waveform.Options) string {
	// Get individual color RGB values to generate a string
	r, g, b, _ := options.BackgroundColor.RGBA()
	bgColorKey := fmt.Sprintf("%d%d%d", r, g, b)

	r, g, b, _ = options.ForegroundColor.RGBA()
	fgColorKey := fmt.Sprintf("%d%d%d", r, g, b)

	r, g, b, _ = options.AlternateColor.RGBA()
	altColorKey := fmt.Sprintf("%d%d%d", r, g, b)

	// Return cache key
	return fmt.Sprintf("%d_%d_%d_%s_%s_%s_%d_%d_%d_%d", id, sizeX, sizeY, bgColorKey, fgColorKey, altColorKey,
		options.Resolution, options.ScaleX, options.ScaleY, options.Sharpness)
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
