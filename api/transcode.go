package api

import (
	"database/sql"
	"fmt"
	"log"
	"net/http"
	"strconv"
	"strings"

	"github.com/mdlayher/wavepipe/data"
	"github.com/mdlayher/wavepipe/transcode"

	"github.com/gorilla/context"
	"github.com/gorilla/mux"
	"github.com/unrolled/render"
)

// GetTranscode returns a transcoded media file stream from wavepipe.  On success, this API will
// return a binary transcode. On failure, it will return a JSON error.
func GetTranscode(w http.ResponseWriter, r *http.Request) {
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
		ren.JSON(w, 400, errRes(400, "no integer transcode ID provided"))
		return
	}

	// Verify valid integer ID
	id, err := strconv.Atoi(pID)
	if err != nil {
		ren.JSON(w, 400, errRes(400, "invalid integer transcode ID"))
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

	// Check for an input codec
	query := r.URL.Query()
	codec := strings.ToUpper(query.Get("codec"))
	if codec == "" {
		// Default to MP3
		codec = "MP3"
	}

	// Check for an input quality
	quality := query.Get("quality")
	if quality == "" {
		// Default to 192kbps
		quality = "192"
	}

	// Create a transcoder using factory
	transcoder, err := transcode.Factory(codec, quality)
	if err != nil {
		// Check for client errors
		switch err {
		// Invalid codec selected
		case transcode.ErrInvalidCodec:
			ren.JSON(w, 400, errRes(400, "invalid transcoder codec: "+codec))
			return
		// Invalid quality for codec
		case transcode.ErrInvalidQuality:
			ren.JSON(w, 400, errRes(400, "invalid quality for codec "+codec+": "+quality))
			return
		// Transcoding subsystem disabled
		case transcode.ErrTranscodingDisabled:
			ren.JSON(w, 503, errRes(503, "ffmpeg not found, transcoding disabled"))
			return
		// MP3 transcoding disabled
		case transcode.ErrMP3Disabled:
			ren.JSON(w, 503, errRes(503, "ffmpeg codec "+transcode.FFmpegMP3Codec+" not found, MP3 transcoding disabled"))
			return
		// OGG transcoding disabled
		case transcode.ErrOGGDisabled:
			ren.JSON(w, 503, errRes(503, "ffmpeg codec "+transcode.FFmpegOGGCodec+" not found, OGG transcoding disabled"))
			return
		// OPUS transcoding disabled
		case transcode.ErrOPUSDisabled:
			ren.JSON(w, 503, errRes(503, "ffmpeg codec "+transcode.FFmpegOPUSCodec+" not found, OPUS transcoding disabled"))
			return
		// All other errors
		default:
			log.Println(err)
			ren.JSON(w, 500, serverErr)
			return
		}
	}

	// Start the transcoder, grab output stream
	transcodeStream, err := transcoder.Start(song)
	if err != nil {
		log.Println(err)
		ren.JSON(w, 500, serverErr)
		return
	}

	// Output the command ffmpeg will use to create the transcode
	log.Println("transcode: command:", transcoder.Command())

	// Now that ffmpeg has started, we must assume binary data is being transferred,
	// so no more error JSON may be sent.

	// Generate a string used for logging this operation
	opStr := fmt.Sprintf("[#%05d] %s - %s [%s %dkbps -> %s %s]", song.ID, song.Artist, song.Title,
		data.CodecMap[song.FileTypeID], song.Bitrate, transcoder.Codec(), transcoder.Quality())

	// Attempt to send transcoded file stream over HTTP
	log.Println("transcode: starting:", opStr)

	// Send transcode stream, no size for now (estimate later)
	if err := HTTPStream(song, transcoder.MIMEType(), -1, transcodeStream, r, w); err != nil {
		// Check for client reset
		if strings.Contains(err.Error(), "connection reset by peer") || strings.Contains(err.Error(), "broken pipe") {
			return
		}

		// Check for cannot seek error, since transcodes cannot currently take advantage of seeking
		if err == ErrCannotSeek {
			// We can send JSON HTTP 416 error, because no data is written on this error
			ren.JSON(w, 416, errRes(416, "seeking is unavailable on transcoded media"))
			return
		}

		log.Println("transcode: error:", err)
		return
	}

	// Wait for ffmpeg to exit
	if err := transcoder.Wait(); err != nil {
		log.Println(err)
		return
	}

	log.Println("transcode: completed:", opStr)
	return
}
