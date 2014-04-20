package core

import (
	"io/ioutil"
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

	// Verify that ffmpeg is available for transcoding
	path, err := exec.LookPath("ffmpeg")
	if err != nil {
		log.Println("transcode: cannot find ffmpeg, transcoding will be disabled")
		return
	}

	// Set ffmpeg location, enable transcoding
	log.Println("transcode: found ffmpeg:", path)
	transcode.Enabled = true
	transcode.FFmpegPath = path

	// Check for codecs which wavepipe uses that ffmpeg is able to use
	ffmpeg := exec.Command(path, "-loglevel", "quiet", "-codecs")
	stdout, err := ffmpeg.StdoutPipe()

	// Start ffmpeg to retrieve its codecs, wait for it to finish
	err2 := ffmpeg.Start()
	codecs, err3 := ioutil.ReadAll(stdout)
	defer stdout.Close()
	err4 := ffmpeg.Wait()

	// Check errors
	if err != nil || err2 != nil || err3 != nil || err4 != nil {
		log.Println("transcode: could not detect ffmpeg codecs, transcoding will be disabled")
		return
	}

	// Check for MP3 transcoding codec
	codecStr := string(codecs)
	if strings.Contains(codecStr, transcode.FFMpegMP3Codec) {
		log.Println("transcode: found", transcode.FFMpegMP3Codec, ", enabling MP3 transcoding")
		transcode.CodecSet.Add("MP3")
	} else {
		log.Println("transcode: could not find", transcode.FFMpegMP3Codec, ", disabling MP3 transcoding")
	}

	// Check for OGG transcoding codec
	if strings.Contains(codecStr, transcode.FFMpegOGGCodec) {
		log.Println("transcode: found", transcode.FFMpegOGGCodec, ", enabling OGG transcoding")
		transcode.CodecSet.Add("OGG")
	} else {
		log.Println("transcode: could not find", transcode.FFMpegOGGCodec, ", disabling OGG transcoding")
	}
}
