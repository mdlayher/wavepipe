package core

import (
	"log"
	"os/exec"
	"strings"

	"github.com/mdlayher/wavepipe/transcode"
)

// transcodeManager manages active file transcodes, their caching, etc, and communicates back
// and forth with the manager goroutine
func transcodeManager(transcodeKillChan chan struct{}) {
	log.Println("transcode: starting...")

	// Perform setup routines for ffmpeg transcoding
	go ffmpegSetup()

	// Trigger events via channel
	for {
		select {
		// Stop transcode manager
		case <-transcodeKillChan:
			// Inform manager that shutdown is complete
			log.Println("transcode: stopped!")
			transcodeKillChan <- struct{}{}
			return
		}
	}
}

// ffmpegSetup performs setup routines for ffmpeg transcoding
func ffmpegSetup() {
	// Disable transcoding until ffmpeg is available
	transcode.Enabled = false

	// Verify that ffmpeg or avconv is available for transcoding
	pathF, errF := exec.LookPath("ffmpeg")
	pathA, errA := exec.LookPath("avconv")
	if errF != nil && errA != nil {
		log.Println("transcode: cannot find ffmpeg/avconv, transcoding will be disabled")
		return
	}

	// Set path based upon if ffmpeg or avconv were found
	var path string
	if errF == nil {
		path = pathF
	} else if errA == nil {
		path = pathA
	}

	// Set ffmpeg location, enable transcoding
	log.Println("transcode: found ffmpeg/avconv:", path)
	transcode.Enabled = true
	transcode.FFmpegPath = path

	// Check for codecs which wavepipe uses that ffmpeg is able to use
	codecs, err := exec.Command(path, "-loglevel", "quiet", "-codecs").Output()
	if err != nil {
		log.Println("transcode: could not detect ffmpeg/avconv codecs, transcoding will be disabled")
		return
	}

	// Check for available codecs
	codecStr := string(codecs)
	for _, c := range []string{transcode.FFmpegMP3Codec, transcode.FFmpegOGGCodec, transcode.FFmpegOPUSCodec} {
		// See if codec is found in output
		if strings.Contains(codecStr, c) {
			log.Println("transcode:", c, "found, enabling transcoding")
			transcode.CodecSet.Add(c)
		} else {
			log.Println("transcode:", c, "not found, disabling transcoding")
		}
	}
}
